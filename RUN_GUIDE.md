# ApiHub 杩愯鎸囧崡

## 榛樿璁块棶鍦板潃
- 绠＄悊鍚庡彴: http://localhost:9011
- OpenAI 鍏煎鎺ュ彛: http://localhost:9011/v1
- Anthropic 鍏煎鎺ュ彛: http://localhost:9011/anthropic/v1

## 鐜瑕佹眰
- Go 1.22+
- Node.js 20+
- npm 10+

## 棣栨鏋勫缓骞惰繍琛?### Windows
```powershell
Copy-Item .env.example .env -Force
cd web
npm ci
npm run build
cd ..
go build -o apihub.exe ./cmd/server
.\apihub.exe
```

### Linux / macOS
```bash
cp .env.example .env
cd web
npm ci
npm run build
cd ..
go build -o apihub ./cmd/server
./apihub
```

## 寮€鍙戞ā寮?### 鍚庣
```bash
go run ./cmd/server
```

### 鍓嶇
```bash
cd web
npm ci
npm run dev
```

鍓嶇寮€鍙戞湇鍔″櫒浼氭妸 `/admin`銆乣/v1`銆乣/anthropic` 浠ｇ悊鍒?`http://localhost:9011`銆?
## 涓€閿惎鍔ㄨ剼鏈?### Windows
```powershell
.\start.bat
```

### Linux / macOS
```bash
chmod +x ./start.sh
./start.sh
```

## Docker
```bash
docker compose -f docker/docker-compose.yml up --build -d
```

## 鍏抽敭閰嶇疆
缂栬緫鏍圭洰褰?`.env`:

```env
SERVER_PORT=9011
DATABASE_PATH=./data/apihub.db
ADMIN_PASSWORD=admin123
SYNC_INTERVAL=60
LOG_LEVEL=info
```

## 楠岃瘉
- 鎵撳紑 `http://localhost:9011`
- 棣栨鍚姩浼氳嚜鍔ㄥ垱寤?`./data/apihub.db`
- 榛樿绠＄悊鍛樿处鍙? `admin`
- 榛樿绠＄悊鍛樺瘑鐮? `admin123`

