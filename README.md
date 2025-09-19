# Desafio Rate Limiter
Objetivo: Desenvolver um rate limiter em Go que possa ser configurado para limitar o número máximo de requisições por segundo com base em um endereço IP específico ou em um token de acesso.

### 🚀 Funcionalidades
✅ Limitação de requisições por endereço IP

✅ Limitação de requisições por token de acesso (header API_KEY)

✅ Configuração via variáveis de ambiente

✅ Persistência em Redis para armazenamento distribuído

✅ Respostas HTTP 429 adequadas quando o limite é excedido

✅ Precedência de token sobre IP (quando ambos estão presentes)

### Como rodar o projeto em ambiente de desenvolvimento
Pré-requisitos:

- Docker e Docker Compose instalados

- Git instalado

Clone e configuração do projeto:

```bash
git clone https://github.com/angelicalombas/rate-limiter
cd rate-limiter
```

### ⚙️ Configuração
Edite o arquivo ```.env``` para personalizar:
- ```REDIS_URL```: URL do Redis (padrão: localhost:6379)

- ```REDIS_PASSWORD```: Senha do Redis (opcional)

- ```RATE_LIMIT_IP```: Limite de requisições por IP (padrão: 5)

- ```RATE_LIMIT_TOKEN```: Limite de requisições por token (padrão: 10)

- ```BLOCK_TIME```: Tempo de bloqueio em segundos (padrão: 100)

- ```ENABLE_IP_LIMIT```: Habilitar limitação por IP (padrão: true)

- ```ENABLE_TOKEN_LIMIT```: Habilitar limitação por token (padrão: true)

- ```USE_MEMORY_LIMITER```: Usar limiter em memória (para testes)

### Execução com Docker:

```bash
docker-compose up --build
```

### Exemplos de Uso:

Requisição normal: ```GET http://localhost:8080/```

Requisição com token: ```GET http://localhost:8080/``` com header ```API_KEY: seu-token```

### Exemplos de Resposta
Sucesso (HTTP 200)
```json
{
  "message": "Request successful",
  "timestamp": 1700000000
}
```
Limite Excedido (HTTP 429)
```json
{
  "error": "you have reached the maximum number of requests or actions allowed within a certain time frame",
  "retry_after": 300.0
}
```


### Testes:

```bash
go test ./tests/ -v
```




### Funcionamento
O rate limiter verifica primeiro se há um token no header API_KEY. Se presente e a limitação por token estiver habilitada, usa o limite configurado para tokens. Caso contrário, usa a limitação por IP.

As chaves no Redis são armazenadas como:

- ```count:ip:<IP>``` para contagem de requisições por IP

- ```count:token:<TOKEN>``` para contagem de requisições por token

- ```block:ip:<IP>``` para bloqueios por IP

- ```block:token:<TOKEN>``` para bloqueios por token
