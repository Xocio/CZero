#ifndef CZERO_FILE_LOCK_H
#define CZERO_FILE_LOCK_H
#include <fcntl.h>
#include <sys/file.h>
#include <unistd.h>

namespace czero {

class FileLock {
    int fd_ = -1;
public:
    explicit FileLock(const char* path) {
        fd_ = open(path, O_CREAT | O_RDWR | O_CLOEXEC, 0644);
        if (fd_ >= 0) flock(fd_, LOCK_EX);
    }
    ~FileLock() {
        if (fd_ >= 0) { flock(fd_, LOCK_UN); close(fd_); }
    }
    FileLock(const FileLock&) = delete;
    FileLock& operator=(const FileLock&) = delete;
};
inline const char* kBasisLock = "/data/adb/modules/CZero/basis/.basis.lock";

}

#endif
