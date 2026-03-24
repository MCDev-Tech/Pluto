# API Docs

**Available Mappings: official, yarn**

### `/`

Ping This server

### `/api/mapping/search`

### Speed Limit

5 times per 2s

#### Queries

- `version`: Target MC version
- `type`: Target mapping type
- `keyword`: Searching keyword
- `translate`: (Optional) Translate to target mapping

### `/api/source/decompile`

### Speed Limit

2 times per 10s

#### Queries

- `version`: Target MC version
- `type`: Target mapping type

### `/api/source/get`

### Speed Limit

5 times per 2s

#### Queries

- `version`: Target MC version
- `type`: Target mapping type
- `class`: Target class, Example: net/minecraft/core/Registry
