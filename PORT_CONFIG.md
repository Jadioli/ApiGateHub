# ApiHub 绔彛閰嶇疆璇存槑

## 鍗犵敤鐨勭鍙?
ApiHub 鏈嶅姟**鍙崰鐢?1 涓鍙?*锛岀敤浜庢彁渚涙墍鏈?HTTP API 鏈嶅姟锛?
- **榛樿绔彛**: 8080
- **鍙厤缃?*: 閫氳繃 `.env` 鏂囦欢淇敼

## 淇敼绔彛鐨勬柟娉?
### 鏂规硶涓€锛氫慨鏀?.env 鏂囦欢锛堟帹鑽愶級

1. 缂栬緫椤圭洰鏍圭洰褰曠殑 `.env` 鏂囦欢
2. 淇敼 `SERVER_PORT` 鐨勫€硷細

```env
SERVER_PORT=9011
```

3. 淇濆瓨鍚庨噸鏂板惎鍔ㄦ湇鍔?
### 鏂规硶浜岋細浣跨敤鐜鍙橀噺

**Windows (PowerShell):**
```powershell
$env:SERVER_PORT="9011"
./apihub.exe
```

**Windows (CMD):**
```cmd
set SERVER_PORT=9011
apihub.exe
```

**Linux/Mac:**
```bash
export SERVER_PORT=9011
./apihub.exe
```

### 鏂规硶涓夛細涓存椂鎸囧畾锛堝崟娆¤繍琛岋級

**Windows (PowerShell):**
```powershell
$env:SERVER_PORT="9011"; ./apihub.exe
```

**Linux/Mac:**
```bash
SERVER_PORT=9011 ./apihub.exe
```

## 绔彛鍐茬獊瑙ｅ喅

### 鏌ョ湅绔彛鍗犵敤鎯呭喌

**Windows:**
```powershell
netstat -ano | findstr :8080
```

**Linux/Mac:**
```bash
lsof -i :8080
# 鎴?netstat -tuln | grep 8080
```

### 鍋滄鍗犵敤绔彛鐨勮繘绋?
**Windows:**
```powershell
# 鏌ユ壘杩涚▼ ID
netstat -ano | findstr :8080

# 鍋滄杩涚▼锛堟浛鎹?PID 涓哄疄闄呰繘绋?ID锛?taskkill /PID <PID> /F
```

**Linux/Mac:**
```bash
# 鏌ユ壘骞跺仠姝㈣繘绋?kill -9 $(lsof -t -i:8080)
```

## 褰撳墠閰嶇疆

鏍规嵁浣犵殑 `.env` 鏂囦欢锛屽綋鍓嶉厤缃负锛?
```
绔彛: 9011
璁块棶鍦板潃: http://localhost:9011
```

## 鎵€鏈夊彲鐢ㄧ殑绔偣

鏈嶅姟鍚姩鍚庯紝鎵€鏈?API 閮介€氳繃鍚屼竴涓鍙ｈ闂細

### 绠＄悊 API
- `http://localhost:9011/admin/login` - 绠＄悊鍛樼櫥褰?- `http://localhost:9011/admin/providers` - Provider 绠＄悊
- `http://localhost:9011/admin/model-configs` - 妯″瀷閰嶇疆鏂规绠＄悊
- `http://localhost:9011/admin/apikeys` - APIKey 绠＄悊
- `http://localhost:9011/admin/logs` - 鏃ュ織鏌ヨ

### 浠ｇ悊 API
- `http://localhost:9011/v1/chat/completions` - OpenAI 鍗忚
- `http://localhost:9011/v1/completions` - OpenAI 鍗忚
- `http://localhost:9011/v1/models` - OpenAI 鍗忚
- `http://localhost:9011/anthropic/v1/messages` - Anthropic 鍗忚
- `http://localhost:9011/anthropic/v1/models` - Anthropic 鍗忚

### 鍓嶇锛堝鏋滄湁锛?- `http://localhost:9011/` - 绠＄悊鐣岄潰

## 娴嬭瘯杩炴帴

鍚姩鏈嶅姟鍚庯紝鍙互浣跨敤浠ヤ笅鍛戒护娴嬭瘯锛?
```bash
curl http://localhost:9011/admin/login \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"password":"admin123"}'
```

濡傛灉杩斿洖 token锛岃鏄庢湇鍔℃甯歌繍琛屻€?
## 闃茬伀澧欓厤缃?
濡傛灉闇€瑕佷粠鍏朵粬鏈哄櫒璁块棶锛岀‘淇濋槻鐏鍏佽璇ョ鍙ｏ細

**Windows 闃茬伀澧?**
```powershell
netsh advfirewall firewall add rule name="ApiHub" dir=in action=allow protocol=TCP localport=9011
```

**Linux (ufw):**
```bash
sudo ufw allow 9011/tcp
```

**Linux (firewalld):**
```bash
sudo firewall-cmd --permanent --add-port=9011/tcp
sudo firewall-cmd --reload
```

## 鐢熶骇鐜寤鸿

1. **浣跨敤鍙嶅悜浠ｇ悊**: 寤鸿浣跨敤 Nginx 鎴?Caddy 浣滀负鍙嶅悜浠ｇ悊
2. **HTTPS**: 閰嶇疆 SSL 璇佷功
3. **淇敼榛樿瀵嗙爜**: 淇敼 `.env` 涓殑 `ADMIN_PASSWORD`
4. **淇敼绠＄悊瀵嗙爜**: 淇敼 `ADMIN_PASSWORD`
5. **闄愬埗璁块棶**: 浣跨敤闃茬伀澧欓檺鍒跺彧鍏佽鐗瑰畾 IP 璁块棶绠＄悊绔彛


