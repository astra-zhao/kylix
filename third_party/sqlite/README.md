# SQLite amalgamation

Android 目标的 `db` 模块不链宿主的 `-lsqlite3`。链接时编译这里的 `sqlite3.c`（SQLite amalgamation），打进产物。

文件不入库。拉取：

```bash
bash scripts/fetch_sqlite_amalgamation.sh
```

或把已有的 `sqlite3.c` 路径放进 `KYLIX_SQLITE_SRC`。

iOS 使用 SDK 里的 `libsqlite3.tbd`，不需要这份源码。
