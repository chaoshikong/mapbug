# mapbug
登记一个go语言中使用map的bug

本项目是在写gofiber项目的时候发现的一个BUG，就是会登记客户端上传的IP地址及当时的时间的功能时发现的map中存在问题

运行本项目后，浏览器多次输入http://127.0.0.1:2178/dns/strings，strings表示不同长度的字符串，多输几次后

再到浏览器中用http://127.0.0.1:2178/list 查看的时候，发现提交的多次数据中，有的数据错掉了

比如我提交了"abcde"，然后又提交了"efg"，用list查询的时候，发现原来提交的"abcde"变成"efgde"了

已提交了这个BUG，地址在https://github.com/golang/go/issues/59917 待回复中
