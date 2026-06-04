# csirender

`csirender` — это расширяемый движок декларативного 2D-рендеринга, написанный на Go. Он позволяет описывать динамические визуальные дашборды с помощью YAML или JSON и рендерить их в стандартные форматы (PNG, BMP) или сырые данные для EPD-экранов (e-ink).

## Возможности

- **Декларативная конфигурация**: Создавайте визуальные элементы (текст, графики, карточки значений, разделители, изображения, геометрические фигуры) прямо в YAML или JSON файлах.
- **Динамическая стилизация**: Встроенная поддержка пороговых значений (Thresholds) для условного изменения цветов, прозрачности и даже путей к картинкам на основе телеметрии.
- **Поддержка изображений**: Рендеринг статических или динамических картинок (PNG, JPEG, GIF, TIFF) с поддержкой прозрачности и режимов масштабирования (fit, stretch, center).
- **Графические примитивы**: Отрисовка прямоугольников, кругов, эллипсов и многоугольников с динамическим вращением, привязанным к телеметрии (например, для стрелки компаса).
- **Расширяемая архитектура**: Реализуйте интерфейс `CustomRenderer` для отрисовки ваших собственных, специфичных для домена компонентов.
- **Оптимизированная загрузка**: Обобщенный `Parser[T]` умеет прозрачно загружать и кешировать конфиги, отслеживая время изменения файлов на диске для предотвращения лишнего парсинга.

## Базовое использование

### 1. Определение конфигурации

`csirender` спроектирован для встраивания в ваши бизнес-приложения. Вы можете парсить конфигурацию напрямую, либо использовать встроенный механизм кеширования.
Полноценные примеры конфигураций и реализации **клиент-серверной архитектуры** находятся в папке [`res/`](res/).

```go
package main

import (
    "github.com/ABespalov/csirender"
    "fmt"
)

// AppConfig встраивает csirender.LayoutConfig и добавляет свои поля бизнес-логики
type AppConfig struct {
    csirender.LayoutConfig `yaml:",inline" json:",inline"`
    MyBusinessLogic string `yaml:"my_logic" json:"my_logic"`
}

var parser = csirender.NewParser[*AppConfig]()

func main() {
    cfg, err := parser.Parse("layout.yaml")
    if err != nil {
        panic(err)
    }
    
    fmt.Println("Ширина экрана:", cfg.Screen.Width)
    fmt.Println("Поле бизнес-логики:", cfg.MyBusinessLogic)
}
```

### 2. Рендеринг интерфейса

Создайте движок (`Engine`), передайте ему вашу телеметрию и вызовите `Render`:

```go
engine := csirender.New()

data := csirender.RenderData{
    Values: map[string]interface{}{
        "temperature": 22.5,
    },
    Charts: make(map[string][]float64),
}

img, err := engine.Render(&cfg.LayoutConfig, data)
if err != nil {
    panic(err)
}

// Сохраните или выведите изображение (img) на экран
```

## Система плагинов

Если вам нужен кастомный тип элемента (например, «спидометр» или «прогресс-бар»), вы можете зарегистрировать свой рендерер:

```go
type MyGaugeRenderer struct {}

func (r *MyGaugeRenderer) RenderElement(dc *gg.Context, rc *csirender.RenderContext, el csirender.Element) error {
    genericEl := el.(*csirender.GenericElement)
    // Читаем произвольные параметры из genericEl.Raw
    // Рисуем спидометр с помощью контекста dc (gg.Context)
    return nil
}

// Регистрация в движке
engine.RegisterRenderer("gauge", &MyGaugeRenderer{})
```

## Лицензия

Copyright (c) 2026, Anton Bespalov.
