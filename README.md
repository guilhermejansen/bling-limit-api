---

# Bling Limit API


Bling Limit API é uma solução robusta para monitorar e limitar as requisições feitas a sua API, prevenindo o abuso e garantindo uma distribuição equitativa dos recursos. Ideal para sistemas que necessitam de um controle refinado sobre o tráfego de entrada.

## Funcionalidades

- **Limite de Requisições:** Impõe um limite na quantidade de requisições que podem ser feitas por um determinado código dentro de um intervalo de tempo.
- **Notificações de Violação:** Envio de notificações sempre que o limite de requisições for excedido, permitindo ações imediatas.
- **Suporte a Docker Swarm:** Facilmente escalável com suporte a Docker Swarm para lidar com altas cargas de tráfego.
- **Integração com Redis:** Utiliza Redis para gerenciar estados e contadores de requisições de forma eficiente.

## Começando

Para começar a utilizar a Bling Limit API, siga os passos abaixo para configurar o ambiente e rodar a aplicação.

### Pré-requisitos

- Docker e Docker Compose
- Uma instância do Redis acessível
- Go (caso deseje compilar localmente)

### Configuração

1. Clone o repositório para sua máquina local:

```bash
git clone https://github.com/seu-usuario/bling_limit.git
cd bling_limit
```

2. Certifique-se de que a instância do Redis esteja rodando e acessível pelas configurações definidas na aplicação.

3. Defina as variáveis de ambiente necessárias. Você pode fazer isso criando um arquivo `.env` na raiz do projeto:

```
REDIS_ADDR=redis_host:6379
REDIS_DB=0
MAX_ATTEMPTS=3
EXPIRE_TIME_SECONDS=5
BLOCK_TIME_MINUTES=1
CALLBACK_ENDPOINT=http://example.com/callback
```

Substitua `redis_host` pelo endereço do seu Redis e o URL pelo endpoint de callback desejado.

### Rodando com Docker

1. Construa a imagem Docker:

```bash
docker build -t bling_limit .
```

2. Inicialize o Docker Swarm (caso ainda não esteja inicializado):

```bash
docker swarm init
```

3. Utilize o `docker stack` para implantar a aplicação:

```bash
docker stack deploy -c docker-compose.yml bling_stack
```

## Uso

Para utilizar a API, envie uma requisição para o endpoint `/api/filter` com o payload desejado. A API verificará se o limite de requisições foi excedido e responderá adequadamente.

### Exemplo de Requisição

```bash
curl -X POST http://localhost:5000/api/filter -d '{"codigo": "ABC123", "outrosDados": "valor"}'
```

## Licença

Distribuído sob a licença MIT. Veja `LICENSE` para mais informações.

## Contato

Meu Portifólio - [Guilherme Jansen -Web Developer](https://guilhermejansen.com.br)

<div align="center">
    <a href="https://www.paypal.com/ncp/payment/K7YAM48FD4Y3Y" target="_blank">
        <img src="https://www.paypalobjects.com/pt_BR/BR/i/btn/btn_donateCC_LG.gif" border="0" alt="Donate with PayPal">
    </a>
    <br>
    <a href="https://www.paypal.com/ncp/payment/K7YAM48FD4Y3Y" target="_blank">
        <img src="https://github.com/guilhermejansen/whaticket_deep_cleaning/raw/main/coffee-qrcode.png" alt="Coffee QR Code">
    </a>
</div>
