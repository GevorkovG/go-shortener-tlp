# go-musthave-shortener-tpl

Шаблон репозитория для трека «Сервис сокращения URL».

## Начало работы

1. Склонируйте репозиторий в любую подходящую директорию на вашем компьютере.
2. В корне репозитория выполните команду `go mod init <name>` (где `<name>` — адрес вашего репозитория на GitHub без префикса `https://`) для создания модуля.

## Обновление шаблона

Чтобы иметь возможность получать обновления автотестов и других частей шаблона, выполните команду:

```
git remote add -m main template https://github.com/Yandex-Practicum/go-musthave-shortener-tpl.git
```

Для обновления кода автотестов выполните команду:

```
git fetch template && git checkout template/main .github
```

Затем добавьте полученные изменения в свой репозиторий.

## Запуск автотестов

Для успешного запуска автотестов называйте ветки `iter<number>`, где `<number>` — порядковый номер инкремента. Например, в ветке с названием `iter4` запустятся автотесты для инкрементов с первого по четвёртый.

При мёрже ветки с инкрементом в основную ветку `main` будут запускаться все автотесты.

Подробнее про локальный и автоматический запуск читайте в [README автотестов](https://github.com/Yandex-Practicum/go-autotests).


//______F____R____O____M____BASH



curl -v -c cookies.txt -X POST -d "https://example.com" http://localhost:8080/

curl -v -b cookies.txt -X POST -d "https://example3.com" http://localhost:8080/

curl -v -b cookies.txt -X GET http://localhost:8080/MIwqRu4sGnu

curl -v -b cookies.txt -X DELETE -H "Content-Type: application/json" -d '["HHSYDQr6Rxl", "MIwqRu4sGnu"]' http://localhost:8080/api/user/urls

curl -v -b cookies.txt http://localhost:8080/api/user/urls

curl -v -b cookies.txt -X DELETE -H "Content-Type: application/json" -d '["HHSYDQr6Rxl", "MIwqRu4sGnu"]' http://localhost:8080/api/user/urls


