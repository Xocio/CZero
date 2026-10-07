<p align="center">
  <img src="assets/logo.png" width="120" alt="CZero logo">
</p>

<h1 align="center">CZero</h1>

<p align="center">Стремимся сделать лучшее решение для очистки кэша и мусора на Android.</p>

<p align="center">
  <a href="https://github.com/Xocio/CZero/releases"><img src="https://img.shields.io/github/v/release/Xocio/CZero?label=release&color=orange" alt="release"></a>
  <a href="https://czeropage.top/en/"><img src="https://img.shields.io/badge/docs-website-blue" alt="docs"></a>
  <img src="https://img.shields.io/badge/root-Magisk%20%7C%20KernelSU%20%7C%20APatch-red" alt="root">
  <img src="https://img.shields.io/badge/%D0%B1%D0%B5%D0%B7%20root-Shizuku-2E7D32" alt="shizuku">
  <a href="https://t.me/CZeroRelease"><img src="https://img.shields.io/badge/Telegram-%D0%9A%D0%B0%D0%BD%D0%B0%D0%BB-26A5E4?logo=telegram&logoColor=white" alt="Telegram Channel"></a>
</p>

<p align="center"><a href="README.md">简体中文</a> · <a href="README_en.md">English</a> · <b>Русский</b></p>

---

CZero — решение для очистки Android: очищает кэш часто используемых приложений, а также выполняет подавление фона, точечное подавление, очистку пустых папок, сортировку файлов, сборку мусора F2FS и fstrim.

Постоянно работающей службы нет — все задачи запускает предельно лёгкий нативный планировщик по `config.json`, а изменения настроек применяются сразу. Повседневное управление выполняется через нативное приложение-компаньон **CZeroX**.

## Две версии

| | **Версия Root** | **Версия Shizuku** |
|---|---|---|
| Получение прав | Magisk / KernelSU / APatch | Shizuku |
| Состав | Модуль CZero + CZeroX (встроен в модуль, устанавливается при прошивке) | Только CZero; программы очистки поставляются вместе с приложением |
| Возможности | Все функции | Без точечного подавления и раздела «Накопитель»; сборку мусора F2FS выполняет системное обслуживание в режиме простоя |

Обе версии используют один идентификатор приложения и не могут быть установлены одновременно. Полное сравнение — в разделе [Choosing an Edition](https://czeropage.top/en/guide/editions) (на английском).

## Сообщество

- **Канал релизов**: [Release](https://t.me/CZeroRelease) — обновления и объявления.
- **Чат**: [Organize](https://t.me/+lwNKCHw_NktjODRh) — отзывы и обсуждение.

## Документация

> **Официальная документация [DOCS](https://czeropage.top/en/)** (на английском и китайском)
>
> Перед использованием настоятельно рекомендуется прочитать документацию; при возникновении проблем сначала загляните в [FAQ](https://czeropage.top/en/guide/faq).

## Возможности

- **Целевая очистка кэша** — отдельные очистители для WeChat / QQ / Douyin, включая двойные приложения; перед очисткой проверяется, не находится ли приложение или игра на переднем плане, а между двумя очистками одного приложения выдерживается минимальный интервал.
- **Подавление фона** — периодически завершает фоновые подпроцессы целевых приложений; основной процесс и push-уведомления не затрагиваются.
- **Точечное подавление** (версия Root, BETA) — приостанавливает фоновые приложения, не завершая их: состояние сохраняется, и возврат к приложению происходит мгновенно. Есть встроенная проверка работы.
- **Сортировка файлов** — раскладывает файлы из папок загрузок и подобных мест по категориям в зависимости от типа; есть история сортировки и полное восстановление.
- **Сборка мусора F2FS и fstrim** — выполняются только тогда, когда устройство действительно простаивает, сохраняя свободное место и скорость записи в долгосрочной перспективе.
- **Пользовательские правила и источники правил** — задавайте собственные пути очистки и белый список или подпишитесь на сторонние источники правил, которые обновляются ежедневно.
- **Корзина** — очищенные файлы сначала перемещаются в корзину, по умолчанию хранятся 7 дней и восстанавливаются на прежнее место.
- **Ассистент Zero** — сканирование накопителя, оценка результата очистки и управление правилами и подавлением в формате диалога; любое изменение выполняется только после подтверждения.
- **Горячая перезагрузка настроек** — изменения применяются сразу; повреждённая конфигурация никогда не заменяет последний рабочий набор задач.

## Загрузка и установка

Последнюю версию можно получить на [странице загрузки](https://czeropage.top/en/download-home) или в [Releases](https://github.com/Xocio/CZero/releases).

| Файл | Описание |
|---|---|
| `CZero_<версия>.zip` | Модуль версии Root со встроенным CZeroX версии Root |
| `CZeroX_Root_<версия>.apk` | Приложение версии Root; нужно, только если оно не установилось автоматически при прошивке |
| `CZero_Shizuku_<версия>.apk` | Приложение версии Shizuku |

**Версия Root**

1. Прошейте zip-архив модуля в Magisk / KernelSU / APatch и с помощью клавиш громкости выберите язык и необходимость сохранить прежнюю конфигурацию.
2. Перезагрузите устройство; CZeroX будет установлен вместе с модулем.

**Версия Shizuku**

1. Установите и запустите [Shizuku](https://shizuku.rikka.app/download/).
2. Установите CZeroX (версия Shizuku), откройте приложение и предоставьте доступ по запросу; всё необходимое приложение развернёт самостоятельно.

## CZeroX

<table>
<tr>
<td valign="top" width="50%">

Нативное приложение-компаньон на Jetpack Compose в стиле [Miuix](https://compose-miuix-ui.github.io/miuix/), доступно на упрощённом китайском, английском и русском языках.

</td>
<td align="center" width="50%">
<img src="assets/webx.png" width="220" alt="Главная страница CZeroX">
</td>
</tr>
</table>

## Star History

<a href="https://www.star-history.com/?repos=Xocio%2FCZero&type=timeline&logscale=&legend=top-left">
 <picture>
   <source media="(prefers-color-scheme: dark)" srcset="https://api.star-history.com/chart?repos=Xocio/CZero&type=timeline&theme=dark&logscale&legend=top-left&sealed_token=twVK7kU7SjickXCW34YQxO2BJE8Ll27rIB3db1HiNE9oyq1tMAXVJy3TiSVIlrdDuAeF0VGVZEdJTbr2bIBoyyvYERJyDzdmRNbeOOwKSMJZRyid1w3R1pxSIclT5LPro3oFtNGwvcdokYqwmWLAIVDeIo_axyrSqJsR1o8BY-_KOHqAIEWhs6lAn4fa" />
   <source media="(prefers-color-scheme: light)" srcset="https://api.star-history.com/chart?repos=Xocio/CZero&type=timeline&logscale&legend=top-left&sealed_token=twVK7kU7SjickXCW34YQxO2BJE8Ll27rIB3db1HiNE9oyq1tMAXVJy3TiSVIlrdDuAeF0VGVZEdJTbr2bIBoyyvYERJyDzdmRNbeOOwKSMJZRyid1w3R1pxSIclT5LPro3oFtNGwvcdokYqwmWLAIVDeIo_axyrSqJsR1o8BY-_KOHqAIEWhs6lAn4fa" />
   <img alt="Star History Chart" src="https://api.star-history.com/chart?repos=Xocio/CZero&type=timeline&logscale&legend=top-left&sealed_token=twVK7kU7SjickXCW34YQxO2BJE8Ll27rIB3db1HiNE9oyq1tMAXVJy3TiSVIlrdDuAeF0VGVZEdJTbr2bIBoyyvYERJyDzdmRNbeOOwKSMJZRyid1w3R1pxSIclT5LPro3oFtNGwvcdokYqwmWLAIVDeIo_axyrSqJsR1o8BY-_KOHqAIEWhs6lAn4fa" />
 </picture>
</a>

## Лицензия

[GNU General Public License v3.0](LICENSE)
