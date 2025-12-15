<div align="center">

# 🔬 Clinical Trials Microservice

### Conectando a comunidade tetraplégica a estudos clínicos que podem mudar vidas

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go)](https://go.dev)
[![API](https://img.shields.io/badge/ClinicalTrials.gov-API_v2-00A86B?style=for-the-badge)](https://clinicaltrials.gov/data-api)
[![Open Source](https://img.shields.io/badge/100%25-Open_Source-00A86B?style=for-the-badge&logo=github)](https://github.com/fcavalcantirj/clinical-trials-microservice)

---

**Parte do ecossistema [SomosTetra](https://github.com/fcavalcantirj/somostetra.org)** | [![Live](https://img.shields.io/badge/🌐_Live-somostetra.org-00A86B?style=flat-square)](https://somostetra.org)

</div>

---

## 🎯 Nossa Missão

Este microserviço foi criado para **conectar pessoas com lesão medular a estudos clínicos** que podem transformar suas vidas. Integrado à plataforma [SomosTetra](https://somostetra.org), ele permite que membros da comunidade tetraplégica encontrem pesquisas relevantes de forma fácil e rápida.

> *"Cada estudo clínico encontrado pode ser a oportunidade que alguém estava esperando."*

---

## ⭐ Features

| Feature | Descrição |
|---------|-----------|
| 🔍 **Busca Completa** | Query clinical trials com múltiplos filtros (condições, status, fase, localização, idade) |
| ⚡ **Rápido & Eficiente** | Construído em Go para alta performance e baixa latência |
| 🔄 **Cache Inteligente** | Cache em memória para reduzir chamadas à API e melhorar tempo de resposta |
| 🛡️ **Rate Limiting** | Rate limiting integrado respeitando limites da ClinicalTrials.gov (50 req/min) |
| 🌍 **Busca por Localização** | Busca geográfica por distância para encontrar estudos próximos |
| 📊 **Dados Ricos** | Retorna informações completas: elegibilidade, locais, contatos e mais |
| 🔌 **API RESTful** | API REST limpa com endpoints GET e POST |

---

## 🔗 Ecossistema SomosTetra

Este microserviço é parte de um projeto maior para ajudar a comunidade tetraplégica brasileira:

| Projeto | Descrição |
|---------|-----------|
| [**somostetra.org**](https://github.com/fcavalcantirj/somostetra.org) | Plataforma principal - conecta a comunidade a estudos clínicos, realiza desejos e amplifica sua voz |
| **clinical-trials-microservice** | Este repositório - API de busca de estudos clínicos |

> 💡 Este microserviço alimenta a funcionalidade de busca de estudos clínicos em [somostetra.org](https://somostetra.org)

---

## 🚀 Quick Start

### Pré-requisitos

- Go 1.21 ou superior
- Docker (opcional, para deploy containerizado)

### Instalação

```bash
# Clone o repositório
git clone https://github.com/fcavalcantirj/clinical-trials-microservice.git
cd clinical-trials-microservice

# Download das dependências
go mod download

# Executar o servidor
go run cmd/server/main.go
```

O servidor estará disponível em `http://localhost:8080`

### Deploy com Docker

```bash
docker build -t clinical-trials-service .
docker run -p 8080:8080 clinical-trials-service
```

---

## 📡 API Reference

### Base URL
```
http://localhost:8080
```

### Endpoints

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| `GET` | `/health` | Health check |
| `GET` | `/api/v1/trials/search` | Buscar trials com query parameters |
| `POST` | `/api/v1/trials/search` | Buscar trials com JSON body |
| `GET` | `/api/v1/trials/{nct_id}` | Buscar trial por NCT ID |

### Filtros Disponíveis

| Parâmetro | Tipo | Descrição | Exemplo |
|-----------|------|-----------|---------|
| `conditions` | string | Condições médicas (separadas por vírgula) | `spinal+cord+injury,tetraplegia` |
| `status` | string | Status do trial | `RECRUITING,NOT_YET_RECRUITING` |
| `phase` | string | Fases do trial | `PHASE2,PHASE3` |
| `latitude` / `longitude` | float | Busca por localização | `34.0522`, `-118.2437` |
| `distance` | integer | Distância em milhas | `50` |
| `minimum_age` / `maximum_age` | string | Filtro por idade | `18 Years`, `65 Years` |
| `page_size` | integer | Resultados por página (max: 1000) | `100` |

### Exemplo Rápido

```bash
# Buscar trials em recrutamento
curl "http://localhost:8080/api/v1/trials/search?status=RECRUITING&page_size=5"

# Buscar por localização (São Paulo)
curl "http://localhost:8080/api/v1/trials/search?latitude=-23.5505&longitude=-46.6333&distance=50"

# Busca complexa via POST
curl -X POST http://localhost:8080/api/v1/trials/search \
  -H "Content-Type: application/json" \
  -d '{
    "conditions": ["spinal cord injury", "tetraplegia"],
    "status": ["RECRUITING"],
    "phase": ["PHASE2", "PHASE3"],
    "page_size": 10
  }'
```

---

## 🏗️ Arquitetura

```
clinical-trials-microservice/
├── cmd/
│   └── server/
│       └── main.go          # Entry point
├── internal/
│   ├── api/
│   │   └── clinicaltrials.go  # ClinicalTrials.gov API client
│   ├── cache/
│   │   └── cache.go           # Caching layer
│   ├── handlers/
│   │   └── trials.go          # HTTP handlers
│   └── models/
│       └── trial.go           # Data models
├── scripts/
│   ├── test_api.sh            # Bash test script
│   └── test_api.py            # Python test script
└── research/
    └── trials_API_integration_guide_for_spinal_cord_injury.md
```

---

## 📦 Response Structure

```json
{
  "trials": [
    {
      "nct_id": "NCT06511934",
      "title": "Feasibility of the BrainGate2 Neural Interface System...",
      "status": "RECRUITING",
      "phase": ["NA"],
      "conditions": ["Tetraplegia", "Spinal Cord Injuries"],
      "locations": [
        {
          "city": "Boston",
          "state": "Massachusetts",
          "country": "United States"
        }
      ],
      "eligibility": {
        "minimum_age": "18 Years",
        "maximum_age": "80 Years",
        "gender": "ALL"
      },
      "sponsor": { "name": "...", "type": "OTHER" },
      "contacts": [{ "name": "...", "email": "..." }],
      "start_date": "2024-07-22",
      "url": "https://clinicaltrials.gov/study/NCT06511934"
    }
  ],
  "total_count": 499,
  "next_page_token": "...",
  "page_size": 10
}
```

---

## ⚙️ Configuração

### Command Line Flags

| Flag | Descrição | Default |
|------|-----------|---------|
| `-port` | Porta do servidor | `8080` |
| `-cache` | Habilitar cache | `true` |
| `-cache-ttl` | TTL do cache | `6h` |

### Deploy na Nuvem

Plataformas recomendadas:
- **Render** (Free tier) - Auto-deploy via `render.yaml`
- **Railway** - Docker com free tier
- **Fly.io** - Deploy global

Ver [DEPLOYMENT.md](./DEPLOYMENT.md) para instruções detalhadas.

---

## 🔬 Comportamento Padrão

Quando nenhum parâmetro `conditions` ou `query` é fornecido, o serviço automaticamente busca por:
- `spinal cord injury OR quadriplegia OR tetraplegia OR paraplegia`

Isso garante que estudos relacionados a SCI sejam encontrados mesmo sem termos de busca explícitos.

---

## 📊 Performance

| Métrica | Valor |
|---------|-------|
| Tempo de resposta (cache) | < 1s |
| Tempo de resposta (API) | 2-5s |
| Rate limit | 50 req/min |
| Cache TTL padrão | 6 horas |

---

## 🧪 Testes

```bash
# Bash test script
./scripts/test_api.sh

# Python tests
python3 scripts/test_api.py

# Go unit tests
go test ./internal/api/...
```

---

## 📚 Referências

- [ClinicalTrials.gov API v2 Documentation](https://clinicaltrials.gov/data-api)
- [Research Guide](./research/trials_API_integration_guide_for_spinal_cord_injury.md)
- [SomosTetra Platform](https://somostetra.org)

---

## 🤝 Contribuindo

1. Fork o repositório
2. Crie uma feature branch
3. Faça suas mudanças
4. Adicione testes
5. Envie um Pull Request

---

<div align="center">

## 🌟 Parte do Ecossistema SomosTetra

**Este microserviço alimenta a busca de estudos clínicos em [somostetra.org](https://somostetra.org)**

[![SomosTetra](https://img.shields.io/badge/🌐_Plataforma_Principal-somostetra.org-00A86B?style=for-the-badge)](https://github.com/fcavalcantirj/somostetra.org)

---

**Feito com ❤️ para a comunidade tetraplégica brasileira**

</div>
