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
