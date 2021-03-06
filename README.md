Консольная программа на Go для генерации сигнатуры указанного файла.

Сигнатура генерируется следующим образом: исходный файл делится на блоки равной (фиксированной) длины (если размер файла не кратен размеру блока, последний фрагмент может быть меньше или дополнен нулями до размера полного блока). Для каждого блока вычисляется значение хеш-функции и дописывается в выходной файл-сигнатуру.  

Интерфейс: командная строка, в которой указаны:

 * Путь до входного файла
 * Путь до выходной файла
 * Размер блока (по умолчанию, 1 Мб)

Обязательные требования:

 * Следует максимально оптимизировать скорость работы утилиты с учетом работы в многопроцессорной среде

Допущения:

 * Размер входного файла может быть много больше размера доступной физической памяти (> 4 Гб)
 * Разрешается использовать только стандартную библиотеку языка Go
 * В качестве хеш-функции можно использовать любую хеш-функцию (MD5, CRC и т.д.)
