# Stock Analyzer Backend

Backend en Go para análisis de datos de stocks que obtiene información de una API externa, la almacena en CockroachDB y proporciona endpoints para consultas y análisis. Integra datos adicionales de Finnhub para complementar la información de las empresas.

## Características

- **API REST** con Gin framework
- **Base de datos** CockroachDB para almacenamiento persistente
- **Integración con Finnhub** para datos adicionales de empresas
- **Sistema de scoring** para análisis de inversiones
- **Filtros avanzados** por múltiples criterios
- **CORS configurado** para frontend
- **Migración automática** de base de datos

## Tecnologías

- **Go 1.24.4**
- **Gin** - Framework web
- **Bun** - ORM para Go
- **CockroachDB** - Base de datos distribuida
- **Finnhub API** - Datos de empresas

## Prerrequisitos

- Go 1.24.4 o superior
- CockroachDB (local o en la nube)
- Cuenta en Finnhub (API key gratuita)

## Instalación

### 1. Clonar el repositorio
```bash
git clone https://github.com/Carlosmercg/stock-analyzer.git
cd stock-analyzer
```

### 2. Instalar dependencias
```bash
go mod download
```

### 3. Configurar variables de entorno

Crear un archivo `.env` en la raíz del proyecto:

```env
# Configuración de la base de datos CockroachDB
DB_USER=tu_usuario
DB_PASSWORD=tu_password
DB_HOST=tu_host
DB_PORT=26257
DB_NAME=stock_analyzer

# API externa para datos de stocks
API_URL=https://api.ejemplo.com/stocks
AUTH_HEADER=Bearer tu_token_aqui

# API de Finnhub para datos adicionales
FINNHUB_APIKEY=tu_api_key_de_finnhub
FINNHUB_URL=https://finnhub.io/api/v1/stock/profile2?symbol=%s&token=%s

# Configuración del servidor
PORT=8085
```

### 4. Ejecutar el proyecto

```bash
# ejecutar directamente
go run main.go
```

El servidor estará disponible en `http://localhost:8085`

## Estructura de Datos


## 🔌 Endpoints Disponibles

### Stocks
- `GET /api/stocks/` - Obtener todos los stocks con paginación
- `GET /api/stocks/filter` - Filtrar stocks con múltiples criterios
- `GET /api/stocks/top` - Top 20 stocks con mejor scoring
- `GET /api/stocks/top-by-brokerage` - Top stocks por corredora
- `GET /api/stocks/brokerages` - Lista de corredoras disponibles
- `GET /api/stocks/ratings` - Lista de ratings disponibles
- `GET /api/stocks/company/info` - Información de empresa desde Finnhub


#### Paginación (`/api/stocks/`)
- `page` - Número de página (default: 1)
- `limit` - Elementos por página (default: 20)

#### Información de Empresa (`/api/stocks/company/info`)
- `ticker` - Símbolo de la empresa (requerido)

#### Top por Corredora (`/api/stocks/top-by-brokerage`)
- `brokerage` - Nombre de la corredora (requerido)

## 🎯 Sistema de Scoring

El sistema calcula un score basado en:

1. **Crecimiento del precio objetivo**: `(target_to - target_from) / target_from * 100`
2. **Rating**: +10 puntos para "Buy" o "Outperform"
3. **Acción**: 
   - +5 puntos para "raised"
   - +2 puntos para "initiated"
   - -5 puntos para "downgraded"

## 🔧 Variables de Entorno Requeridas

| Variable | Descripción | Ejemplo |
|----------|-------------|---------|
| `DB_USER` | Usuario de CockroachDB | `root` |
| `DB_PASSWORD` | Contraseña de CockroachDB | `tu_password` |
| `DB_HOST` | Host de CockroachDB | `localhost` |
| `DB_PORT` | Puerto de CockroachDB | `26257` |
| `DB_NAME` | Nombre de la base de datos | `stock_analyzer` |
| `API_URL` | URL de la API de stocks | `https://api.ejemplo.com/stocks` |
| `AUTH_HEADER` | Header de autorización | `Bearer tu_token` |
| `FINNHUB_APIKEY` | API key de Finnhub | `tu_api_key` |
| `FINNHUB_URL` | URL template de Finnhub | `https://finnhub.io/api/v1/stock/profile2?symbol=%s&token=%s` |
| `PORT` | Puerto del servidor | `8080` |


## 📞 Contacto

Carlos Mercado - [@Carlosmercg](https://github.com/Carlosmercg)

Link del proyecto: [https://github.com/Carlosmercg/stock-analyzer](https://github.com/Carlosmercg/stock-analyzer)
