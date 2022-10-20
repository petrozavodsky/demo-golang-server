
Старт proxy `go run main.go -type proxy -port=900 `

Старт сервисов 

`go run main.go -port=8081`

`go run main.go -port=8081`

можно указать другие порты если захочется

Очистка хранилища 

`curl -X DELETE --location "http://localhost:9000/flush"  -H "Content-Type: application/json"`
