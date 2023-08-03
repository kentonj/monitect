import io

def main():
    b = io.BufferedRWPair(io.BytesIO(), io.BytesIO())
    for i in range(10):
        b.write(bytes(f'hello there: {i}', 'utf-8'))
    print(b.peek(10))

    pass

if __name__ == '__main__':
    main()
