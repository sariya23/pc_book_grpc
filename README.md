# PC book

PC book - это RPC/REST сервис, который позволяет хранить тех. хар-ки ноутбуков и искать их по фильтрам. Создание сущностей доступно только пользователям с ролью admin. Авторизация и аутентификация производится по JWT. Проксирование REST запроса реализовано через [gRPC-Gateway](https://github.com/grpc-ecosystem/grpc-gateway). 

## API references

Взаимодействие с сервисами возможно как по REST, так и по RPC. 

REST API doc:

- laptop service - https://app.swaggerhub.com/apis/sariya/PC-book-laprop-service/1.0. Взаимодейсвтие с сущностями
- auth service - https://app.swaggerhub.com/apis/sariya/PC-book-auth-service/1.0. Сервис аутентификации

### TODO

- PostgreSQL
- Docker compose run
