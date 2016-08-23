# batch-批量下载图片
## Introduction
用于并发下载图片
最大图片并发下载限制为50张

batch中所有的下载操作动作，取自程序中维护的goroutine票池

该票池保证最大并发下载不超过50，以此防止因为过多的并发用户请求造成内存泄露，以及打开过多的文件句柄

## TODO
目前仅实现了下载到本地，并通过php脚本转存

## Usage
```
Usage: batch [options...]

Options:
    -d string 临时文件存放目录（default "/tmp")
    -f string 执行脚本路径,这里需要根据实际情况配置(defuault "/home/john/save_img.php")

```

## Start

`go build && ./batch`