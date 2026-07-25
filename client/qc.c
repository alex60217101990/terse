/* qdf-hookc: minimal AF_UNIX client for the qdf-hook daemon.
 * Usage: qdf-hookc <socket-path>   (request on stdin, reply on stdout)
 *
 * The daemon reads a request until EOF (no framing), so we half-close the
 * write side with shutdown(SHUT_WR) before reading the reply. Any error exits
 * non-zero so the hook's shell `|| qdf-hook post` fallback runs.
 */
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <sys/socket.h>
#include <sys/un.h>

int main(int argc, char **argv) {
    if (argc != 2) return 2;

    int fd = socket(AF_UNIX, SOCK_STREAM, 0);
    if (fd < 0) return 1;

    struct sockaddr_un addr;
    memset(&addr, 0, sizeof(addr));
    addr.sun_family = AF_UNIX;
    if (strlen(argv[1]) >= sizeof(addr.sun_path)) { close(fd); return 1; }
    strcpy(addr.sun_path, argv[1]);

    if (connect(fd, (struct sockaddr *)&addr, sizeof(addr)) < 0) { close(fd); return 1; }

    char buf[65536];
    ssize_t n;
    while ((n = read(STDIN_FILENO, buf, sizeof(buf))) > 0) {
        ssize_t off = 0;
        while (off < n) {
            ssize_t w = write(fd, buf + off, (size_t)(n - off));
            if (w < 0) { close(fd); return 1; }
            off += w;
        }
    }
    if (n < 0) { close(fd); return 1; }

    if (shutdown(fd, SHUT_WR) < 0) { close(fd); return 1; }

    while ((n = read(fd, buf, sizeof(buf))) > 0) {
        ssize_t off = 0;
        while (off < n) {
            ssize_t w = write(STDOUT_FILENO, buf + off, (size_t)(n - off));
            if (w < 0) { close(fd); return 1; }
            off += w;
        }
    }
    close(fd);
    return n < 0 ? 1 : 0;
}
