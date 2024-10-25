def seq(input_data):
    input_data = input_data.strip()

    ASCII_LOWER_START = 97
    ASCII_LOWER_END = 122

    concat = ""
    for char in input_data:
        inChar = ASCII_LOWER_START
        while True:
            print('\"' + concat + chr(inChar) + '\"')
            if inChar == ord(char):
                break
            inChar += 1
        concat += char

if __name__ == "__main__":
    import sys
    input_data = sys.stdin.read()
    seq(input_data)
