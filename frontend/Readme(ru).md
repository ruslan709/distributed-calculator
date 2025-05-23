## Интерфейсная часть для “Распределенного калькулятора”
## Описание
Это веб-интерфейс для распределенного калькулятора с поддержкой регистрации, авторизации, отправки выражений для расчета. Интерфейс написан на чистом HTML, CSS и JavaScript, сборка не требуется.
## Особенности
* Регистрация и логин пользователя (JWT)
* Отправка арифметических выражений для расчета
* Очистка истории расчетов
* Просмотр состояния серверов (orchestrator и калькуляторов)
* Автоматическое обновление результатов и статусов серверов
## Структура проекта
```text
frontend/
├── index.html      # Домашняя страница приложения
├── styles.css      # Основные стили
└── script.js       # Сценарии приложений
```
## Запуск
> **Переместить в папку frontend**
> Обычно это папка `frontend` или аналогичная.
> **Откройте файл `index.html` в своем браузере**
> Вы можете просто дважды щелкнуть по файлу или открыть его через контекстное меню `Открыть с помощью".
> **Рекомендуется**
> Для корректной работы запросов на выборку используйте локальный сервер (например, Live Server для VSCode или любой http-сервер).
> **Убедитесь, что серверные службы запущены**
## Использование
***Регистрация и вход в систему***
- Введите свое имя пользователя и пароль.
- Нажмите на ссылку “Зарегистрироваться”, чтобы зарегистрироваться.
- После успешной регистрации войдите в систему.
***Введите выражение***
- Введите выражение (например, `2+2*3") в поле “Калькулятор”.
- Нажмите “Рассчитать”.
***Просмотрите результаты***
- Как только выражение будет отправлено, оно появится в списке с уникальным идентификатором и статусом (ожидание или результат).
- Нажмите “Обновить результаты”, чтобы обновить статусы.
***Очистка истории***
- Кнопка “Очистить все вычисления” удаляет все вычисления пользователя.
## Пример запроса API
* `POST /api/v1/login` - вход пользователя в систему
* `POST /api/v1/register` - регистрация пользователя
* `POST /submit-вычисление" - отправка выражения
* `ПОЛУЧАТЬ /get-вычисления-по-пользователю?userId=...` - история вычислений
* `ПОЛУЧАТЬ /get-calculation-result?id=...` - результат по идентификатору
* `ОПУБЛИКОВАТЬ /очистить-все-вычисления" - очистка истории
* `GET /orchestrator-status" - статус оркестратора
* `GET /ping-servers` - статусы калькулятора

Переведено с помощью DeepL.com (бесплатная версия)
## Интерфейс
## Интерфейс
![Иллюстрация для проекта](https://github.com/ruslan709/distributed-calculator/tree/main/frontend/interface)
