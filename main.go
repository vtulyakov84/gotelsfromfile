package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Config содержит настройки программы
type Config struct {
	InputFile  string
	OutputFile string
	Unique     bool
	Sort       bool
	Verbose    bool
}

func main() {
	// Определяем флаги
	config := parseFlags()

	// Выполняем основную логику
	phones, err := extractPhones(config)
	if err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		os.Exit(1)
	}

	// Выводим результаты
	if err := outputResults(phones, config); err != nil {
		fmt.Printf("Ошибка: %v\n", err)
		os.Exit(1)
	}
}

// parseFlags парсит аргументы командной строки
func parseFlags() Config {
	var config Config

	// Определяем флаги
	flag.StringVar(&config.InputFile, "file", "", "Путь к входному файлу (обязательный)")
	flag.StringVar(&config.InputFile, "f", "", "Путь к входному файлу (сокращенный)")
	flag.StringVar(&config.OutputFile, "output", "", "Путь к выходному файлу (опционально)")
	flag.StringVar(&config.OutputFile, "o", "", "Путь к выходному файлу (сокращенный)")
	flag.BoolVar(&config.Unique, "unique", false, "Показывать только уникальные номера")
	flag.BoolVar(&config.Unique, "u", false, "Показывать только уникальные номера (сокращенный)")
	flag.BoolVar(&config.Sort, "sort", false, "Сортировать номера")
	flag.BoolVar(&config.Sort, "s", false, "Сортировать номера (сокращенный)")
	flag.BoolVar(&config.Verbose, "verbose", false, "Подробный вывод")
	flag.BoolVar(&config.Verbose, "v", false, "Подробный вывод (сокращенный)")

	// Кастомное описание использования

	programName := filepath.Base(os.Args[0])

	flag.Usage = func() {
		fmt.Printf("Использование: %s [опции] -file <путь_к_файлу>\n\n", os.Args[0])
		fmt.Println("Опции:")
		flag.PrintDefaults()
		fmt.Println("\nПримеры:")
		fmt.Printf("  %s -file input.txt\n", programName)
		fmt.Printf("  %s -f input.txt -o output.txt -unique -sort\n", programName)
		fmt.Printf("  %s --file=input.txt\n", programName)
	}

	flag.Parse()

	// Проверяем обязательный флаг
	if config.InputFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	return config
}

// extractPhones извлекает телефонные номера из файла
func extractPhones(config Config) ([]string, error) {
	// Открываем файл
	file, err := os.Open(config.InputFile)
	if err != nil {
		return nil, fmt.Errorf("не удалось открыть файл %s: %v", config.InputFile, err)
	}
	defer file.Close()

	// Регулярное выражение для поиска 10 цифр подряд
	// Ищем последовательности из 10 цифр, перед которыми может быть 7 или 8, или не быть
	// phoneRegex := regexp.MustCompile(`[78]?\d{10}`)
	phoneRegex := regexp.MustCompile(`[78]?\d{3}\D?\d{3}\D?\d{2}\D?\d{2}`)

	scanner := bufio.NewScanner(file)

	// Используем map для уникальных номеров
	phoneMap := make(map[string]bool)
	var phones []string

	if config.Verbose {
		fmt.Printf("Чтение файла: %s\n", config.InputFile)
		fmt.Printf("Уникальные номера: %v\n", config.Unique)
		fmt.Printf("Сортировка: %v\n", config.Sort)
		fmt.Println()
	}

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Ищем номера в строке
		// Сначала пробуем найти номера с разделителями
		matches := phoneRegex.FindAllString(line, -1)

		for _, match := range matches {
			// Извлекаем только цифры
			digits := extractDigits(match)

			// Проверяем длину
			if len(digits) < 10 {
				continue
			}

			// Берем последние 10 цифр (убираем +7 или 8)
			var phone string
			if len(digits) > 10 {
				phone = digits[len(digits)-10:]
			} else {
				phone = digits
			}

			// Проверяем, что номер состоит из 10 цифр
			if len(phone) != 10 {
				continue
			}

			// Добавляем номер
			addPhone(phone, &phones, phoneMap, config.Unique)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("ошибка чтения файла: %v", err)
	}

	if config.Verbose {
		fmt.Printf("Всего найдено номеров: %d\n", len(phones))
		if config.Unique {
			fmt.Printf("Уникальных номеров: %d\n", len(phones))
		}
	}

	// Сортировка
	if config.Sort {
		sort.Strings(phones)
		if config.Verbose {
			fmt.Println("Номера отсортированы")
		}
	}

	return phones, nil
}

// addPhone добавляет номер в список с учетом уникальности
func addPhone(phone string, phones *[]string, phoneMap map[string]bool, unique bool) {
	if unique {
		if !phoneMap[phone] {
			phoneMap[phone] = true
			*phones = append(*phones, phone)
		}
	} else {
		*phones = append(*phones, phone)
	}
}

// extractDigits извлекает все цифры из строки
func extractDigits(s string) string {
	var result strings.Builder
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			result.WriteRune(ch)
		}
	}
	return result.String()
}

// outputResults выводит результаты
func outputResults(phones []string, config Config) error {
	if len(phones) == 0 {
		fmt.Println("Телефонные номера не найдены")
		return nil
	}

	// Если указан выходной файл
	if config.OutputFile != "" {
		file, err := os.Create(config.OutputFile)
		if err != nil {
			return fmt.Errorf("не удалось создать файл %s: %v", config.OutputFile, err)
		}
		defer file.Close()

		writer := bufio.NewWriter(file)
		for _, phone := range phones {
			if _, err := writer.WriteString(phone + "\n"); err != nil {
				return fmt.Errorf("ошибка записи в файл: %v", err)
			}
		}
		if err := writer.Flush(); err != nil {
			return fmt.Errorf("ошибка сброса буфера: %v", err)
		}

		fmt.Printf("Результат сохранен в файл: %s\n", config.OutputFile)
		fmt.Printf("Всего номеров: %d\n", len(phones))
		return nil
	}

	// Вывод в консоль
	fmt.Printf("Найдено телефонных номеров: %d\n\n", len(phones))
	for i, phone := range phones {
		fmt.Printf("%d. %s\n", i+1, phone)
	}

	return nil
}
