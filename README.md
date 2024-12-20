# Telegram Shopping Bot

Это бот для Telegram, который позволяет управлять списками покупок. Написан на языке Go, демонстрирует использование Docker и развертывание в облачной среде.

## Функциональность
- Создание и управление списками покупок для каждого чата.
- Добавление, удаление и вычеркивание пунктов списка.
- Поддержка множества пользователей с изолированными списками.

## Предварительные требования
- Установленный [Docker](https://www.docker.com/get-started/).
- Токен бота Telegram от [BotFather](https://core.telegram.org/bots#botfather).
- Установленный [CLI Yandex Cloud](https://yandex.cloud/ru/docs/cli/quickstart) (использовался мной для хранения Docker-образа).
- Облачная VM, например [Cloud.ru](https://cloud.ru/) (для бота достаточно мощностей Free tier)

## Установка
1. Склонируйте репозиторий:
   ```bash
   git clone https://github.com/I-ivlev-I/telegram-shopping-bot.git
   cd telegram-shopping-bot
   
2. Соберите Docker-образ:
   ```bash
   docker build -t telegram-bot .
   
3. Запустите бота:
   ```bash
   docker run -d --name telegram-bot -e TELEGRAM_BOT_TOKEN=ваш-токен telegram-bot

## Запуск бота на виртуальной машине

Эта инструкция объясняет, как запустить бота на виртуальной облачной машине с использованием Docker-образа, размещённого в Yandex.Cloud, и сервиса Cloud.ru.

### 1. Размещение Docker-образа в Yandex.Cloud Container Registry

1. Авторизуйтесь в Yandex.Cloud:
   ```bash
   yc init

2. Создайте реестр контейнеров (если ещё не создан):
   ```bash
	yc container registry create --name telegram-bot-registry

3. Скопируйте ID реестра:
   ```bash
   yc container registry list

4. Авторизуйтесь в реестре:
  ```bash
   echo <Ваш OAuth-токен>  | docker login --username oauth --password-stdin cr.yandex
   
5. Соберите Docker-образ:
  ```bash
  docker build -t telegram-bot .

6. Задайте тег для образа:
  ```bash
  docker tag telegram-bot cr.yandex/<registry_id>/telegram-bot:latest

7. Отправьте образ в реестр:
  ```bash
  docker push cr.yandex/<registry_id>/telegram-bot:latest

Теперь ваш Docker-образ размещён в Yandex.Cloud Container Registry.

### 2. Подготовка виртуальной машины в Cloud.ru

1. Создайте виртуальную машину на Cloud.ru с доступом в интернет 

2. Установите Docker на виртуальной машине:
  ```bash
  sudo apt update
  sudo apt install -y docker.io
  sudo systemctl start docker
  sudo systemctl enable docker

3. Установите CLI Yandex Cloud
Для Linux:
     ```bash
     curl -sSL https://storage.yandexcloud.net/yandexcloud-yc/install.sh | bash
	 
  Затем добавьте CLI в ваш `PATH`:
     ```bash
     export PATH=$HOME/yandex-cloud/bin:$PATH
  Проверьте, что установка выполнена:
     ```bash
     yc version
4. Авторизуйтесь в Yandex.Cloud Container Registry на виртуальной машине:
  ```bash
  echo <Ваш OAuth-токен>  | docker login --username oauth --password-stdin cr.yandex
  
### 3. Запуск контейнера на виртуальной машине

1. Загрузите Docker-образ с Yandex.Cloud Container Registry:
  ```bash
  docker pull cr.yandex/<registry_id>/telegram-bot:latest

2. Запустите контейнер с передачей токена:
  ```bash
  docker run -d --name telegram-bot -e TELEGRAM_BOT_TOKEN=ваш_токен cr.yandex/<registry_id>/telegram-bot:latest

3. Проверьте, что контейнер работает:
  ```bash
  docker logs telegram-bot

4. Для настройки автоматического перезапуска (на случай перезагрузки виртуальной машины) выполните:
  ```bash
  docker update --restart always telegram-bot

Теперь ваш бот успешно работает на облачной виртуальной машине, используя Docker-образ из Yandex.Cloud. 