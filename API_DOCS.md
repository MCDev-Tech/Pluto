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
- `filter`: Filters: 0x100 class, 0x10 method, 0x1 field
- `translate`: (Optional) Translate to target mapping

### `/api/source/get`

### Speed Limit

5 times per 2s

#### Queries

- `version`: Target MC version
- `type`: Target mapping type
- `class`: Target class, Example: net/minecraft/core/Registry
