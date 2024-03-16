import os, io, sys

def main():
    message_to_server = "Hello, Server!\n"
    p2c_read, p2c_write = os.pipe()
    c2p_read, c2p_write = os.pipe()

    child_pid = os.fork()
    if child_pid == 0:
        # Child process
        os.close(p2c_write)
        os.close(c2p_read)
        os.dup2(p2c_read, 0)
        os.dup2(c2p_write, 1)
        os.close(p2c_read)
        os.close(c2p_write)
        os.execlp('nc', 'nc', 'server', '12345')
    else:
        # Parent process
        os.close(p2c_read)
        os.close(c2p_write)
        os.write(p2c_write, message_to_server.encode())
        response = os.read(c2p_read, 1024).decode()
        os.close(p2c_write)
        os.close(c2p_read)
        print("Test passed! ✅" if response == message_to_server else "Test failed! ❌")

if __name__ == '__main__':
    main()