# calc_api_go
Сервис подсчёта арифметических выражений

## Описание
Сервис предоставляет функцию для вычисления арифметических выражений состоящих из односимвольных идентификаторов и знаков арифметических действий. Входящие данные - цифры(рациональные), операции +, -, *, /, операции приоритезации ( и ).

## Инструкция по запуску
1. Клонируйте репозиторий
```bash
git clone https://github.com/RichCake/calc_api_go.git
```
2. Перейдите в корневую директорию проекта
```bash
cd путь/к/проекту
```
3. Запустите файл cmd/main.go
```bash
go run cmd/main.go
```

## Формат запроса и ответа
Чтобы получить результата вычисления выражения отправьте POST запрос на адрес /api/v1/calculate с выражением в ключе expression. Пример запроса:
```
curl --location 'localhost:8080/api/v1/calculate' \
--header 'Content-Type: application/json' \
--data '{
  "expression": "2+2*2"
}'
```
Ответ представляет из себя JSON с ключом result, если нет ошибки, в ином случае с ключом error. Подробнее про ошибки далее. 

Пример ответа:
```
{
    "result": 6
}
```

## Ошибки возвращаемые сервером
Если вы послали на сервер некорректное выражение, то он пришлет ответ в формате `{"error": "какая-то ошибка"}`. Если ошибка заключается в неправильном арифметическом выражении, то сервер вернет код статуса 422 Unprocessable Entity. 

Ошибки, связанные с арифметическим выражением:

| N  | Ошибка                              | Описание                                                                 | Пример curl запроса |
|----|-------------------------------------|--------------------------------------------------------------------------|---------------------|
| 1  | `{"error": "mismatched bracket"}`   | Указывает на неправильную скобочную последовательность.                  | `curl --location 'localhost:8080/api/v1/calculate' --header 'Content-Type: application/json' --data '{"expression": "2+2*2)"}'` |
| 2  | `{"error": "invalid symbols"}`      | Указывает на некорректные символы в выражении. Корректные символы: цифры (рациональные), операции +, -, *, /, операции приоритезации ( и ). | `curl --location 'localhost:8080/api/v1/calculate' --header 'Content-Type: application/json' --data '{"expression": "2+2*a"}'` |
| 3  | `{"error": "invalid operations placement"}` | Указывает на некорректную расстановку арифметических операций. Например, 2++2 или 2*(+2+2). | `curl --location 'localhost:8080/api/v1/calculate' --header 'Content-Type: application/json' --data '{"expression": "2++2"}'` |
| 4  | `{"error": "division by zero"}`     | Возвращается при делении на ноль.                                        | `curl --location 'localhost:8080/api/v1/calculate' --header 'Content-Type: application/json' --data '{"expression": "2/0"}'` |
| 5  | `{"error": "invalid expression"}`   | Возвращается в иных случаях.                                             | Примера не будет, так как большинство ситуаций обработаны в других ошибках |

Другие ошибки:

| N  | Ошибка                              | Описание                                                                 | Пример curl запроса |
|----|-------------------------------------|--------------------------------------------------------------------------|---------------------|
| 1  | `{"error":"missing request body"}`  | Указывает на пустое тело запроса. Код статуса 400 Bad Request.           | `curl --location 'localhost:8080/api/v1/calculate' --header 'Content-Type: application/json' --data ''` |
| 2  | `{"error":"'expression' field is required"}` | Указывает что в теле запроса нет ключа expression или он пустой. Код статуса 400 Bad Request. | `curl --location 'localhost:8080/api/v1/calculate' --header 'Content-Type: application/json' --data '{"expression": ""}'` |
| 3  | `{"error":"method not allowed"}`    | Вызывается если запрос произведен не с методом POST. Код статуса 405 Method Not Allowed. | `curl --location --request GET 'localhost:8080/api/v1/calculate'` |

## Структура проекта
```
calc_api_go
├── cmd
│   └── main.go
├── internal
│   └── application
│       ├── application.go
│       └── ...
└── pkg
    └── calculation
        ├── calculation.go
        └── ...
```

Описание директорий:

`pkg` — пакеты, функционал которых можно будет использовать как внутри этого модуля, так и сторонними модулями

`internal` — пакеты, которые не могут быть использованы другими модулями

`cmd` — пакет main для запуска программы

## Тесты
В проекте предусмотренны тесты для функции вычисления арифметических выражений в файле `pkg/calculation/calculation_test.go` и тесты для сервера в файле `internal/application/application_test.go`.

Чтобы запустить тесты перейдите в корневую директорию проекта и выполните команду
```bash
go test -v ./...
```

## Логи
Логи будут хранится в файле `logs.txt`. В логах содержится информация о работе сервера, ошибках и сбоев. Пример:
```
time=2024-12-21T14:34:15.177+03:00 level=INFO msg="Starting server" port=8080
time=2024-12-21T14:34:21.401+03:00 level=INFO msg="Received request" method=POST path=/api/v1/calculate
time=2024-12-21T14:34:21.402+03:00 level=INFO msg="Get expression" expression=20*(2+7)
time=2024-12-21T14:34:21.402+03:00 level=INFO msg="Calculation result" result=180
```
