/*
Package main предоставляет расширенный статический анализатор Go-кода, объединяющий:
- стандартные анализаторы из golang.org/x/tools/go/analysis/passes
- анализаторы класса SA из staticcheck.io
- выбранные анализаторы из других классов staticcheck.io
- дополнительные публичные анализаторы
- кастомный анализатор noosexit

# Использование

Установка:

	go install ./cmd/staticlint

Базовый запуск:

	staticlint ./...

Анализ конкретного пакета:

	staticlint ./path/to/package

Рекурсивный анализ:

	staticlint ./...

Флаги:

	-json       вывод в формате JSON
	-exclude    список проверок для исключения (через запятую)
	-fix        автоматическое исправление (где возможно)
	-tests      включать тестовые файлы
	-cpu        ограничение количества используемых CPU

# Включенные анализаторы

## Стандартные анализаторы (golang.org/x/tools/go/analysis/passes)
- appends       - проверяет правильность использования append
- asmdecl       - проверяет соответствие ассемблерных деклараций
- assign        - обнаруживает бесполезные присваивания
- atomic        - проверяет использование sync/atomic
- bools         - обнаруживает ошибки в булевых операциях
- buildtag      - проверяет теги сборки
- cgocall       - проверяет вызовы CGO
- composite     - проверяет композитные литералы
- copylock      - проверяет копирование мьютексов
- errorsas      - проверяет использование errors.As
- fieldalignment - обнаруживает неоптимальное выравнивание структур
- httpresponse  - проверяет обработку HTTP ответов
- loopclosure   - обнаруживает проблемы с замыканиями в циклах
- printf        - проверяет форматные строки
- shadow        - обнаруживает затенение переменных
- structtag     - проверяет теги структур
- unusedresult  - проверяет неиспользуемые результаты вызовов

## Staticcheck анализаторы (класс SA)
- SA1000+       - набор проверок безопасности и корректности кода
- SA1019        - проверка устаревших пакетов
- SA4000+       - проверки логических ошибок

## Другие анализаторы
- ineffassign   - обнаруживает неэффективные присваивания
- noosexit      - кастомный анализатор, запрещающий os.Exit в main()

# Кастомный анализатор noosexit

Анализатор запрещает прямой вызов os.Exit в функции main пакета main.

Пример неправильного кода:

	func main() {
	    os.Exit(1) // ошибка
	}

Рекомендуемая замена:

	func main() {
	    if err := run(); err != nil {
	        log.Fatal(err) // правильно
	    }
	}

	func run() error {
	    return nil
	}
*/
package main

import (
	"strings"

	"github.com/gordonklaus/ineffassign/pkg/ineffassign"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/appends"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/atomicalign"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/cgocall"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/analysis/passes/deepequalerrors"
	"golang.org/x/tools/go/analysis/passes/defers"
	"golang.org/x/tools/go/analysis/passes/directive"
	"golang.org/x/tools/go/analysis/passes/errorsas"
	"golang.org/x/tools/go/analysis/passes/fieldalignment"
	"golang.org/x/tools/go/analysis/passes/findcall"
	"golang.org/x/tools/go/analysis/passes/framepointer"
	"golang.org/x/tools/go/analysis/passes/gofix"
	"golang.org/x/tools/go/analysis/passes/hostport"
	"golang.org/x/tools/go/analysis/passes/httpmux"
	"golang.org/x/tools/go/analysis/passes/httpresponse"
	"golang.org/x/tools/go/analysis/passes/ifaceassert"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/analysis/passes/loopclosure"
	"golang.org/x/tools/go/analysis/passes/lostcancel"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/pkgfact"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/reflectvaluecompare"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/shift"
	"golang.org/x/tools/go/analysis/passes/sigchanyzer"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/stdversion"
	"golang.org/x/tools/go/analysis/passes/stringintconv"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/testinggoroutine"
	"golang.org/x/tools/go/analysis/passes/tests"
	"golang.org/x/tools/go/analysis/passes/timeformat"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
	"golang.org/x/tools/go/analysis/passes/unreachable"
	"golang.org/x/tools/go/analysis/passes/unsafeptr"
	"golang.org/x/tools/go/analysis/passes/unusedresult"
	"golang.org/x/tools/go/analysis/passes/unusedwrite"
	"golang.org/x/tools/go/analysis/passes/usesgenerics"
	"golang.org/x/tools/go/analysis/passes/waitgroup"
	"honnef.co/go/tools/analysis/facts/nilness"
	"honnef.co/go/tools/quickfix"
	"honnef.co/go/tools/simple"
	"honnef.co/go/tools/staticcheck"
	"honnef.co/go/tools/stylecheck"

	"github.com/GevorkovG/go-shortener-tlp/cmd/staticlint/myanalizer"
)

func main() {
	var analyzers []*analysis.Analyzer

	// Добавляем все стандартные анализаторы
	analyzers = append(analyzers,
		appends.Analyzer,
		asmdecl.Analyzer,
		atomicalign.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		buildssa.Analyzer,
		buildtag.Analyzer,
		cgocall.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		ctrlflow.Analyzer,
		deepequalerrors.Analyzer,
		defers.Analyzer,
		directive.Analyzer,
		errorsas.Analyzer,
		fieldalignment.Analyzer,
		findcall.Analyzer,
		framepointer.Analyzer,
		gofix.Analyzer,
		hostport.Analyzer,
		httpmux.Analyzer,
		httpresponse.Analyzer,
		ifaceassert.Analyzer,
		inspect.Analyzer,
		loopclosure.Analyzer,
		lostcancel.Analyzer,
		nilfunc.Analyzer,
		nilness.Analysis,
		pkgfact.Analyzer,
		printf.Analyzer,
		reflectvaluecompare.Analyzer,
		shadow.Analyzer,
		shift.Analyzer,
		sigchanyzer.Analyzer,
		stdmethods.Analyzer,
		stdversion.Analyzer,
		stringintconv.Analyzer,
		structtag.Analyzer,
		testinggoroutine.Analyzer,
		tests.Analyzer,
		timeformat.Analyzer,
		unmarshal.Analyzer,
		unreachable.Analyzer,
		unsafeptr.Analyzer,
		unusedresult.Analyzer,
		unusedwrite.Analyzer,
		usesgenerics.Analyzer,
		waitgroup.Analyzer,
	)

	// Добавляем анализаторы staticcheck SA
	for _, v := range staticcheck.Analyzers {
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			analyzers = append(analyzers, v.Analyzer)
		}
	}

	// Добавляем по одному анализатору из других классов
	analyzers = append(analyzers,
		simple.Analyzers[0].Analyzer,
		stylecheck.Analyzers[0].Analyzer,
		quickfix.Analyzers[0].Analyzer,
	)

	// Добавляем дополнительные анализаторы
	analyzers = append(analyzers,
		ineffassign.Analyzer,
	)

	// Добавляем собственный анализатор
	analyzers = append(analyzers, myanalizer.NoOsExitAnalyzer)

	multichecker.Main(analyzers...)
}
