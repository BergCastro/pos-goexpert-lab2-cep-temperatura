# pos-goexpert-labs-cep-temperatura

## Instruções de configuração

- Em service-a copiar .env-example e renomiar para .env
- Em service-a preencher variavel PORT
- Em service-b copiar .env-example e renomiar para .env
- Em service-b preencher variavel PORT e WEATHER_API_KEY(https://www.weatherapi.com/)
- rodar o comando docker-compose up --build

## Instruções de uso do serviço

- Ex.: localhost:8080/zipcode POST body: { "cep": "01001000" }
