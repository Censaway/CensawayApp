# 🛡️ CensawayApp

<p align="center">
  <img src="build/appicon.png" width="128" height="128" alt="CensawayApp Logo">
</p>

<p align="center">
  <b>Современный, быстрый и красивый VPN-клиент с поддержкой VLESS, Reality и TUN-режима.</b>
</p>

<p align="center">
  <img src="https://img.shields.io/github/v/release/Censaway/CensawayApp?style=flat-square&color=purple" alt="Release">
  <img src="https://img.shields.io/badge/Platform-Windows%20%7C%20Linux%20%7C%20macOS-blue?style=flat-square" alt="Platform">
  <img src="https://img.shields.io/badge/Built%20with-Wails%20%2B%20React-00ADD8?style=flat-square" alt="Wails">
  <img src="https://img.shields.io/badge/Core-Sing--Box-white?style=flat-square" alt="Sing-Box">
</p>

---

## ✨ Возможности

CensawayApp — это GUI-оболочка для мощного ядра **sing-box**, написанная на Go (Wails) и React. Приложение создано для обхода блокировок с максимальным комфортом и производительностью.

*   🚀 **Поддержка протоколов:** VLESS (Vision, Reality, TLS, TCP/WS/gRPC/HTTP).
*   🌐 **Режимы работы:**
    *   **TUN Mode:** Виртуальный сетевой интерфейс (маршрутизация всего трафика системы, включая игры и терминал).
    *   **System Proxy:** Автоматическая настройка системного прокси.
*   🧠 **Smart Routing:**
    *   Прямое подключение к российским сайтам (`.ru`, `.rf` и список GeoIP RU) — не замедляет локальный трафик.
    *   Пользовательские правила маршрутизации для доменов, IP и **процессов** (Windows/Linux).
*   🛡️ **Безопасность и Контроль:**
    *   **Проверка IP:** Отображение реального IP-адреса, через который идет трафик, прямо в дашборде.
    *   **Single Instance:** Защита от повторного запуска приложения.
*   📦 **Управление профилями:** Поддержка подписок (Subscriptions) и ручного добавления VLESS-ссылок.
*   🖥️ **Удобство:**
    *   Сворачивание в системный трей.
    *   Автозапуск при старте системы.
    *   Автоподключение к последнему активному серверу.
    *   **Фоновое обновление:** Автоматическая проверка новых версий каждый час.
*   📊 **Мониторинг:** Отображение текущей скорости и задержки (ping) до серверов.
*   🎨 **Интерфейс:** Современный, адаптивный Dark Mode дизайн (Glassmorphism).

## 📥 Установка

Перейдите на вкладку [**Releases**](https://github.com/Censaway/CensawayApp/releases) и скачайте последнюю версию для вашей ОС.

### Windows
1. Скачайте `CensawayApp-amd64-installer.exe`.
2. Запустите установщик.
3. **Права администратора:** Приложение можно запускать как обычный пользователь. Права администратора (UAC) будут запрошены **только** при переключении в режим **TUN Mode**.

### Linux
1. Скачайте бинарный файл `CensawayApp`.
2. Дайте права на выполнение: `chmod +x CensawayApp`.
3. Запустите. Приложение автоматически создаст ярлык в меню приложений.
4. *Примечание:* Для работы требуются `libgtk-3`, `libwebkit2gtk` и `libayatana-appindicator3`.

### macOS
1. Скачайте `CensawayApp-mac.dmg`.
2. Перетащите в папку Applications.
3. *Если появляется ошибка безопасности:* Выполните в терминале `xattr -cr /Applications/CensawayApp.app`.

---

## 🛠️ Сборка из исходников

### Требования
*   **Go** 1.21+
*   **Node.js** 18+ (NPM)
*   **Wails CLI:** `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### Запуск в режиме разработки
```bash
wails dev
```

### Сборка релизной версии

#### Windows (создание установщика)
Требуется установленный **NSIS**.

*На Windows:*
```powershell
wails build -platform windows/amd64 -nsis
```

*На Linux (кросс-компиляция):*
```bash
# Установите mingw-w64 и nsis
sudo apt install mingw-w64 nsis # Debian/Ubuntu
# или
yay -S mingw-w64-gcc nsis # Arch Linux

# Сборка
CC=x86_64-w64-mingw32-gcc \
CXX=x86_64-w64-mingw32-g++ \
CGO_ENABLED=1 \
wails build -platform windows/amd64 -nsis
```

#### Linux
```bash
wails build -platform linux/amd64
```

#### macOS
```bash
wails build -platform darwin/universal
```

---

## 🏗️ Стек технологий

*   **Frontend:** React, TypeScript, TailwindCSS, Zustand, Vite.
*   **Backend:** Go (Golang).
*   **Framework:** [Wails v2](https://wails.io).
*   **VPN Core:** [sing-box](https://github.com/SagerNet/sing-box) (автоматическая загрузка и управление).
*   **Driver (Windows):** Wintun.

## 📄 Лицензия

Этот проект распространяется под лицензией MIT. Подробнее см. в файле [LICENSE](LICENSE).

---

<p align="center">
  Сделано с ❤️ для свободного интернета.
</p>