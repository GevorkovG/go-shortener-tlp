// Анализатор, который проверяет и запрещает
// использование прямых вызовов os.Exit() в функции main() пакета main.
package myanalyzer

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// NoOsExitAnalyzer определяет анализатор, который проверяет и запрещает
// использование прямых вызовов os.Exit() в функции main() пакета main.
//
// Анализатор помогает соблюдать best practices по обработке ошибок,
// рекомендуя использовать log.Fatal() или возврат ошибок вместо os.Exit().
//
// Пример неправильного использования:
//
//	func main() {
//	    os.Exit(1) // вызовет ошибку анализатора
//	}
//
// Рекомендуемая замена:
//
//	func main() {
//	    if err := run(); err != nil {
//	        log.Fatal(err) // правильный способ
//	    }
//	}
var NoOsExitAnalyzer = &analysis.Analyzer{
	Name:     "noosexit",
	Doc:      "запрещает прямой вызов os.Exit в функции main пакета main",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      runNoOsExit,
}

// runNoOsExit реализует основную логику анализатора.
//
// Параметры:
//   - pass: анализ.Pass предоставляет информацию о анализируемом пакете
//
// Возвращает:
//   - interface{}: дополнительные результаты (не используются)
//   - error: ошибка выполнения анализатора
//
// Алгоритм работы:
//  1. Получает AST инспектор из зависимостей
//  2. Фильтрует узлы AST, оставляя только вызовы функций
//  3. Для каждого вызова функции проверяет:
//     - Является ли вызовом os.Exit
//     - Находится ли вызов в функции main() пакета main
//     - Не является ли файл временным файлом сборки
//  4. При обнаружении нарушения создает отчет
func runNoOsExit(pass *analysis.Pass) (interface{}, error) {
	// Получаем инспектор из зависимостей
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Фильтр для узлов AST: только вызовы функций
	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		call := n.(*ast.CallExpr)

		// Проверяем, что это вызов os.Exit
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return
		}

		pkg, ok := sel.X.(*ast.Ident)
		if !ok || pkg.Name != "os" || sel.Sel.Name != "Exit" {
			return
		}

		// Получаем позицию в исходном коде
		pos := pass.Fset.Position(call.Pos())

		// Пропускаем файлы в кэш-директории
		if strings.Contains(pos.Filename, filepath.Join(".cache", "go-build")) {
			return
		}

		// Проверяем, что находимся в функции main пакета main
		if pass.Pkg.Name() == "main" && isInMainFunction(pass.Fset, call) {
			pass.Reportf(
				call.Pos(),
				"прямой вызов os.Exit() запрещен в функции main (найдено в %s:%d:%d)",
				pos.Filename,
				pos.Line,
				pos.Column,
			)
		}
	})

	return nil, nil
}

// isInMainFunction проверяет, находится ли AST-узел внутри функции main.
//
// Параметры:
//   - fset: *token.FileSet для работы с позициями в коде
//   - node: AST-узел для проверки
//
// Возвращает:
//   - bool: true если узел находится в функции main, иначе false
//
// Примечание:
// Функция проверяет только наличие узла внутри объявления функции
// с именем "main", но не проверяет принадлежность к пакету main.
func isInMainFunction(fset *token.FileSet, node ast.Node) bool {
	for _, f := range node.(*ast.File).Decls {
		fn, ok := f.(*ast.FuncDecl)
		if ok && fn.Name.Name == "main" {
			return true
		}
	}
	return false
}
