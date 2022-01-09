# golang-crud-mysql

Приложение представляет собой API на языке Golang с функциями CRUD для MySQL. <br>
Также реализован UI при помощи HTML5, CSS3, немного JQuery.

В БД две таблицы: Работники (Workers) и Бухгалтерия (Bookkeeping).

Обработаны следующие endpoints:
* "/" - метод GET - главная страница. Можно выбрать одну из таблиц для работы с ней.
* "/workers" - метод GET - страница просмотра записей, метод DELETE с передаваемым value, равным id записи, - удаление записи с переданным id.
* "/workers/add" - метод GET - страница с формами для ввода данных, метод POST - добавление записи.
* "/workers/update/" - метод GET - страница с формами для ввода данных, метод POST - добавление записи.
* "/bookkeeping" -  метод GET - страница просмотра записей, метод DELETE с передаваемым value, равным id записи, - удаление записи с переданным id.
* "/bookkeeping/add" - метод GET - страница с формами для ввода данных, метод POST - добавление записи.
* "/bookkeeping/update/ -  метод GET - страница с формами для ввода данных, метод POST - добавление записи.

Для работы с HTTP использует стандартный пакет "net/http", для работы с БД - стандартный пакет "database/sql" с драйвером ["github.com/go-sql-driver/mysql"](https://github.com/go-sql-driver/mysql). <br>
Для логирования используется стандартный пакет "log". <br>
Для создание динамических HTML страниц используются шаблоны из стандартного пакета "html/template". <br>
Для создания стилей используется фреймворк bootstrap5. <br>
