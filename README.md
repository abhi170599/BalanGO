我是光年实验室高级招聘经理。
我在github上访问了你的开源项目，你的代码超赞。你最近有没有在看工作机会，我们在招软件开发工程师，拉钩和BOSS等招聘网站也发布了相关岗位，有公司和职位的详细信息。
我们公司在杭州，业务主要做流量增长，是很多大型互联网公司的流量顾问。公司弹性工作制，福利齐全，发展潜力大，良好的办公环境和学习氛围。
公司官网是http://www.gnlab.com,公司地址是杭州市西湖区古墩路紫金广场B座，若你感兴趣，欢迎与我联系，
电话是0571-88839161，手机号：18668131388，微信号：echo 'bGhsaGxoMTEyNAo='|base64 -D ,静待佳音。如有打扰，还请见谅，祝生活愉快工作顺利。

# **BalanGO**
### An adaptive load balancer developed with Go

######  Balango prioritizes servers based on real time data including active connections,average response time and history of handled connections


### **Features**

 1.  #####   Multiple modes of configuration
                -Round Robin Configuration
                -Adaptive Load Balancing
 2.  #####     Easy to use CLI



### Usage
1.  Generate a **Config.xml** file
```xml     
<?xml version="1.0" encoding="UTF-8"?>
<Config>
  	 <Servers>
           <Server address="http://localhost:5000"/>
           <Server address="http://localhost:5001"/>
           <Server address="http://localhost:5002"/>
     
           <Mode>RR</Mode>
           <Port>3030</Port>
   	</Servers>
 </Config>
```  


2. Use the **balango CLI** to start the load balancer

    > balango  --config=Config.xml     


### **Binaries**	

[Linux Executable][linexe270420]





[linexe270420]: https://github.com/abhi170599/BalanGO/raw/master/build/Linux/balango
[Windows Executable][winexe270420]





[winexe270420]: https://github.com/abhi170599/BalanGO/raw/master/build/Windows/balango
[Darwin Executable][darexe270420]





[darexe270420]: https://github.com/abhi170599/BalanGO/raw/master/build/Darwin/balango
