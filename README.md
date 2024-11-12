# sozu

Итак sozu. Это отпугивалка оленей,
которая поможет вам собрать ваши данные в кучку.

## Базовый функционал

На текущий момент реализовано 2 режима работы: буфер и агрегатор

### Буфер

Буфер накапливает масив данных и отдает его по запросу.

Чтобы создать фабрику буферов, выполните `fabric := sozu.NewBuffer[type]()`
И затем создайте любое количество буферов вызвав
`output, flush := fabric.Create(ctx, input)` где

- `input` канал типа type, куда предпологается передавать входные данные.
- `ctx` контекст для экстреного завершения работы буфера
- `output` канал типа []type с масивом выходных данных
- `flush` функция принуждающая буфер отдать данные

Пример кода с использованием буфера можно посмотреть
[тут](example/buffer/buffer.go)

### Агрегатор

Агрегатор работает похожим образом, но позволяет сэкономить на алокации масива,
в случае если вам необходимо просто выполнить простую операцию с данными.
Например если вам нужно хранить только максимальный элемент.

Чтобы создать фабрику агрегаторов выполните
`fabric := sozu.NewAggregator[type](aggregateFunc)` где
`aggregateFunc` это функция с сигнатурой `func(type, type) type`.
Затем вы можете создать любое количество агрегаторов вызвав
`output, flush := fabric.Create(ctx, input)` где

- `input` канал типа type, куда предпологается передавать входные данные.
- `ctx` контекст для экстреного завершения работы буфера
- `output` канал типа []type с масивом выходных данных
- `flush` функция принуждающая буфер отдать данные

Пример кода с использованием агрегатора можно посмотреть
[тут](example/aggregator/aggregator.go)

:warning: Для коректного завершения работы буфера и агрегатора нужно закрыть
входной канал,
это позволит буферу коректно отдать данные прежде чем завершить работу.
Закрытие контекста привидет к экстреной остановке работы буфера с потерей всех
данных находящихся в нем

## Расширеный функционал

Базовые элементы библиотеки sozu бесполезны, но их функционал можно расширить,
добавляя к ним опции модицицирующие их поведение.
Далее все опции будут демонстрироваться на примере Buffer но они применимы
ко всем базовым элементам sozu.
Опции так же можно сочетать в любых комбинациях и в любом порядке.
Опции применяются в порядке передаче их в конструктор

### Лимит на количество

Опция WithLimit позволяет задать лимит количества значений в буфере.
По достижении этого лимита произойдет принудительный "сброс" буфера.
"Сброс" буфера из другой опции, так же приведет к откату счетчика limit
на соответствующее количество элементов

Создать фабрику буферов с лимитом элементов можно добавив соотвествующую опцию
в параметры конструктора:

```go
fabric := sozu.NewBuffer[type](sozu.WithLimit(limit))
```

где `limit` это максимальнео количество элементов в буфере

Пример кода можно посмотреть [тут](example/limited_buffer/limited_buffer.go)

:warning: WithLimit переопределяет capacity буфера, это стоит учитывать
при использовании нескольких опций влияющих на capacity. Итоговый буфер будет
создан с capacity первой афектящей опции из списка параметров конструктора.

### Лимит по времени

Опция WithTimer позволяет задать лимит времени на хранение данных в буфере.
По истечении времени произойдет принудительный "сброс" буфера.
"Сброс" из другой опции также приведет к сбросу таймера,
при условии что в буфере не осталось элементов.

Создать фабрику буферов с лимитом по времени можно добавив соответствуующую
опцию в параметры конструктора:

```go
fabric := sozu.NewBuffer[type](sozu.WithTimer(duration))
```

где `duration` это максимальное время хранения данных в буфере типа
time.Duration

Пример кода можно посмотреть [тут](example/timer_buffer/timer_buffer.go)

### Флип флоп опция

Опция WithFlipFLop позволяет "сбросить" буфер при изменении какойлибо
характеристики данных

Создать фабрику буферов с FlipFlop опцией можно вызвав

```go
fabric := sozu.NewBuffer[type](sozu.WithFlipFlop(criteria))
```

где criteria это функция с сигнатурой `func[V any, C comparable](V) C`
эта функция будет применятся ко всем входным данным и при изменении ее ответа
произойдет "сброс" буфера

Пример кода можно посмотреть [тут](example/flip_flop_buffer/flip_flop_buffer.go)

### Мультиплексор

Опция WithMultiplexor позволяет независимо хранить данные
с разными характеристиками

Создать буфер с мультиплексором можно вызвав

```go
fabric := sozu.NewBuffer[type](sozu.WithMultiplexor(separator, capacity))
```

где `separator` функция с сигнатурой `func[V any, C comparable](V) C`
а `capacity` ожидаемое количество вариантов ответа этой функции

Под копотом буфер с этой опцией создаст "подбуферы" для значений с разными
ответами функции separator

Пример кода можно посмотреть
[тут](example/multiplexor_buffer/multiplexor_buffer.go)

Комбинируя эту опцию с другими можно добится сложного поведения, как
[тут](example/multiplexor_combination/multiplexor_combination.go)

## TODO
- опция для указания капасити в случае если limit не задан
- опции для ручной постановки колбеков
- опция для сброса по достижению количества определенных элементов
