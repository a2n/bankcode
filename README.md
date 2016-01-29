# 金融機構代碼清單
![臺灣銀行](http://is5.mzstatic.com/image/thumb/Purple6/v4/5e/9b/e2/5e9be257-4e41-cb68-a4b5-88838017896e/source/60x60bb.jpg)
## 用途
程式開發需要臺灣金融機構代碼清單。


## 特色
*  JSON 格式
*  比對 ```Last-Modified``` 資訊。
*  支援 Javascript。

## 來源
玉山銀行 - [ATM跨行轉帳金融機構代號一覽表](http://www.esunbank.com.tw/event/announce/BankCode.htm)

## 更新時間
Mon, 18 Jan 2016 08:37:38 GMT

#用法
## JSON
格式採用 JSON，直接解析使用。
```json
	[
	  {
        "No": "機構代碼",
        "Name": "機構名稱",
        "Type": "機構類別"
      }
    ]
```


## JavaScript
```html
<html>
<head>
	<meta charset="utf-8">
	<title>Title</title>
</head>
<body>
	<script src="https://rawgit.com/a2n/bankcode/master/bankcode.min.js"></script>
	<script>
		(function() {
			Bankcode.forEach(function(bank) {
				console.log(bank.Name + '(' + bank.No + '): ' + bank.Type)
			});
		})();
	</script>
</body>
</html>
```


# TODOs
*  整合 git 自動更新。


# 授權
[GNU General Public License v3.0](https://opensource.org/licenses/GPL-3.0)