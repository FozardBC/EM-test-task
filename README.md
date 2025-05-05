## Для запуска БД:
Необходимо:
- docker
- postgres image

$ docker pull postgres

для установки образа postgres

$ docker compose up -d --remove-orphans

для запуска БД - localhost:5432

## Для запуска приложения:

- Создать в корнеовй папке файл .env
- заполнить его следующимим переменными:

  ```
    #APP
    LOG_MODE=dev #debug | dev

    ### HTTP_SERVER
    SRV_HOST=localhost
    SRV_PORT=8082

    ### DATABASE
    DB_CONN_STRING=postgresql://[username]:[password]@[url]/test-task-db?sslmode=disable
  ```

### С установленым go 

    - $ go mod download
    - $ go build -o ./cmd/test-task ./cmd/test-task 
    - $ ./cmd/test-task/test-task 

### Запуск бинарного файла

    - $ ./bin

### Документация:

- http://urlPath/api/v1/swagger/index.html
