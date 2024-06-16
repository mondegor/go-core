# Описание GoWebCore v0.18.1
Этот репозиторий содержит описание библиотеки GoWebCore.

## Статус библиотеки
Библиотека находится в стадии разработки.

## Описание библиотеки
Библиотека с базовой функциональностью для разработки web сервисов, в которую входят:
- общие интерфейсы, такие как `logger`, `router`, `validator` и другие, которые могут быть реализованы уже в конкретных проектах;
- адаптеры логгеров: стандартного и `rs/zerolog`;
- адаптер стандартного http сервера;
- адаптеры http роутеров:
    - `go-chi/chi/v5`;
    - `julienschmidt/httprouter`;
- адаптер cors (`rs/cors`);
- адаптер валидатора (`go-playground/v10`);
- работа с пользовательскими разрешениями и привилегиями (ролевая модель);
- разграничение доступа к модулям из различных API;
- часто используемые программные, системные и пользовательские ошибки, которые возникают в разных слоях программы;
- пакеты с часто используемыми функциями: генерация токенов, преобразование IP и т.д.;
- парсеры для некоторых типов данных, которые поступают из http запросов;
- парсеры для работы с файлами и изображениями;

## Подключение библиотеки
`go get github.com/mondegor/go-webcore`