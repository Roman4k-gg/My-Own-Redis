import socket
import time
import sys

def send_command(s, cmd_bytes):
    s.sendall(cmd_bytes)
    return s.recv(1024)

def run_tests():
    print("ПОЛНОЕ ТЕСТИРОВАНИЕ КЛОНА REDIS\n")
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    s.connect(("127.0.0.1", 6379))

    print("--- 1. БАЗОВЫЕ КОМАНДЫ ---")
    print("PING:", send_command(s, b"*1\r\n$4\r\nPING\r\n"))
    print("ECHO:", send_command(s, b"*2\r\n$4\r\nECHO\r\n$11\r\nHello World\r\n"))

    print("\n--- 2. ХРАНИЛИЩЕ И УДАЛЕНИЕ ---")
    print("SET user Roman:", send_command(s, b"*3\r\n$3\r\nSET\r\n$4\r\nuser\r\n$5\r\nRoman\r\n"))
    print("EXISTS user:", send_command(s, b"*2\r\n$6\r\nEXISTS\r\n$4\r\nuser\r\n"))
    print("GET user:", send_command(s, b"*2\r\n$3\r\nGET\r\n$4\r\nuser\r\n"))
    print("DEL user:", send_command(s, b"*2\r\n$3\r\nDEL\r\n$4\r\nuser\r\n"))
    print("GET user (after DEL):", send_command(s, b"*2\r\n$3\r\nGET\r\n$4\r\nuser\r\n"))

    print("\n--- 3. ТЕСТ TTL (ЭКСПИРАЦИЯ) ---")
    print("SET temp_key 123 EX 2:", send_command(s, b"*5\r\n$3\r\nSET\r\n$8\r\ntemp_key\r\n$3\r\n123\r\n$2\r\nEX\r\n$1\r\n2\r\n"))
    print("GET temp_key (сразу):", send_command(s, b"*2\r\n$3\r\nGET\r\n$8\r\ntemp_key\r\n"))
    print("⏳ Ждем 3 секунды...")
    time.sleep(3)
    print("GET temp_key (после паузы):", send_command(s, b"*2\r\n$3\r\nGET\r\n$8\r\ntemp_key\r\n"))

    print("\n--- 4. ПОДГОТОВКА К ТЕСТУ AOF (ПЕРСИСТЕНТНОСТЬ) ---")
    print("Записываем важные данные на диск...")
    send_command(s, b"*3\r\n$3\r\nSET\r\n$8\r\nsave_me1\r\n$6\r\nData_1\r\n")
    send_command(s, b"*3\r\n$3\r\nSET\r\n$8\r\nsave_me2\r\n$6\r\nData_2\r\n")
    print("Готово. В базе лежат ключи 'save_me1' и 'save_me2'.")
    
    print("\nЧтобы проверить AOF:")
    print("1. Убей сервер в терминале (Ctrl+C)")
    print("2. Запусти сервер заново (go run cmd/server/main.go)")
    print("3. Запусти этот скрипт с аргументом: python test.py verify")
    s.close()

def verify_aof():
    print("ПРОВЕРКА ВОССТАНОВЛЕНИЯ ДАННЫХ ИЗ AOF\n")
    s = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    try:
        s.connect(("127.0.0.1", 6379))
    except ConnectionRefusedError:
        print("Ошибка: Сервер не запущен! Запусти его сначала.")
        return

    print("GET save_me1:", send_command(s, b"*2\r\n$3\r\nGET\r\n$8\r\nsave_me1\r\n"))
    print("GET save_me2:", send_command(s, b"*2\r\n$3\r\nGET\r\n$8\r\nsave_me2\r\n"))
    s.close()
    print("\nЕсли видно Data_1 и Data_2 — клон Redis работает")

if __name__ == "__main__":
    if len(sys.argv) > 1 and sys.argv[1] == "verify":
        verify_aof()
    else:
        run_tests()
