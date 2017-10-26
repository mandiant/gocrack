import argparse
import time
import sys


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Prints numbers')
    parser.add_argument('--max', type=int, default=5, help='max number to print')
    parser.add_argument('--rc', type=int, default=0, help='return code to exit as')
    args = parser.parse_args()

    i = 0
    while i < args.max:
        sys.stdout.write('I: %d\n' % i)
        sys.stdout.flush()
        i += 1
        time.sleep(2)

    sys.exit(args.rc)
