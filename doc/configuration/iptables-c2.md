# How To: Configure iptables on Ubuntu

Iptables is a firewall, installed by default on all official Ubuntu distributions (Ubuntu, Kubuntu, Xubuntu). When you install Ubuntu, iptables is there, but it allows all traffic by default. For detailed information about iptables, you can visit [ubuntu's help page](https://help.ubuntu.com/community/IptablesHowTo).

Iptables should be set according to project requirements, in the below tutorial we assume that the requirement is to block all traffic expect from known networks. 

!!! info
    running ``` man iptables ``` on a ubuntu server will provide addition documentation, and help understand better the command line arguments. 

1. To list your current rules in iptables, run
```bat
sudo iptables -L -v
```
If you have just set up your server, you will have no rules, and you should see
```bat
Chain INPUT (policy ACCEPT)
target     prot opt source               destination

Chain FORWARD (policy ACCEPT)
target     prot opt source               destination

Chain OUTPUT (policy ACCEPT)
target     prot opt source               destination
```
2. Block all incoming, transiting, and outgoing traffic using the steps below:
```bat
sudo iptables --policy FORWARD DROP
sudo iptables --policy INPUT DROP
sudo iptables --policy OUTPUT DROP
sudo iptables -L -v
```
The output of the last command should look like:
```bat
Chain INPUT (policy DROP 15 packets, 1260 bytes)
target     prot opt source               destination

Chain FORWARD (policy DROP 0 packets, 0 bytes)
target     prot opt source               destination

Chain OUTPUT (policy DROP 302 packets, 18120 bytes)
target     prot opt source               destination
```

!!! info
    In order to block all traffic, you need to be logged directly on the server. Using a remote connection will get you kicked out.
3. Allow all traffic from local network can be achieved by typing:
```bat
sudo iptables -A INPUT -s localnet/24  -j ACCEPT
sudo iptables -A OUTPUT -s localnet/24  -j ACCEPT
sudo iptables -L -v
```
The output of these commands looks like: 
```bat
Chain INPUT (policy DROP 18 packets, 1512 bytes)
 pkts bytes target     prot opt in     out     source               destination         
 2138  181K ACCEPT     all  --  any    any     localnet/24          anywhere            

Chain FORWARD (policy DROP 0 packets, 0 bytes)
 pkts bytes target     prot opt in     out     source               destination         

Chain OUTPUT (policy DROP 302 packets, 18120 bytes)
 pkts bytes target     prot opt in     out     source               destination         
 4990  378K ACCEPT     all  --  any    any     localnet/24          anywhere 
```
!!! info
    If allowing additional networks is required run step 3 again, and replace localnet with any ip address from the network to allow. The output of ``` sudo iptables -L -v``` will reflect those changes accordingly. 

4. If allowing access to the outside world is required when connections are initiated from the server, run the below command:
```bat
sudo iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
sudo iptables -L -v
```
iptables configuration should like this after that: 
```bat
Chain INPUT (policy DROP 18 packets, 1512 bytes)
 pkts bytes target     prot opt in     out     source               destination         
 2138  181K ACCEPT     all  --  any    any     localnet/24          anywhere 
 1084 5731K ACCEPT     all  --  any    any     anywhere             anywhere             state RELATED,ESTABLISHED           
Chain FORWARD (policy DROP 0 packets, 0 bytes)
 pkts bytes target     prot opt in     out     source               destination         

Chain OUTPUT (policy DROP 302 packets, 18120 bytes)
 pkts bytes target     prot opt in     out     source               destination         
 4990  378K ACCEPT     all  --  any    any     localnet/24          anywhere 
```
5. Localhost connections should always be allowed, so local programs can interact between each other using ``` localhost or 127.0.0.1 ```. In order to achieve that run the below commands:

```bat
sudo iptables -A INPUT -s localhost  -j ACCEPT
sudo iptables -A OUTPUT -s localhost  -j ACCEPT
sudo iptables -L -v
```
The output should be close to: 
```bat
Chain INPUT (policy DROP 0 packets, 0 bytes)
 pkts bytes target     prot opt in     out     source               destination         
 2202  185K ACCEPT     all  --  any    any     localnet/24          anywhere            
 1086 5732K ACCEPT     all  --  any    any     anywhere             anywhere             state RELATED,ESTABLISHED
    0     0 ACCEPT     all  --  any    any     localhost            anywhere            
    0     0 ACCEPT     all  --  any    any     localhost            anywhere            

Chain FORWARD (policy DROP 0 packets, 0 bytes)
 pkts bytes target     prot opt in     out     source               destination         

Chain OUTPUT (policy DROP 0 packets, 0 bytes)
 pkts bytes target     prot opt in     out     source               destination         
 5071  388K ACCEPT     all  --  any    any     localnet/24          anywhere            
    0     0 ACCEPT     all  --  any    any     localhost            anywhere            
    0     0 ACCEPT     all  --  any    any     localhost            anywhere  
```
6. In order to save this configuration, we need to install [iptables-persistent](https://packages.ubuntu.com/xenial/admin/iptables-persistent) using the below command:

```bat
sudo apt-get install iptables-persistent
```
Next, save iptables configuration using:
```bat
sudo iptables-save
```
The output should look similar to: 
```bat
# Generated by iptables-save v1.6.0 on Mon Jun 24 19:40:24 2019
*filter
:INPUT DROP [0:0]
:FORWARD DROP [0:0]
:OUTPUT DROP [0:0]
-A INPUT -s x.x.x.0/24 -j ACCEPT
-A INPUT -m state --state RELATED,ESTABLISHED -j ACCEPT
-A INPUT -s 127.0.0.1/32 -j ACCEPT
-A INPUT -s 127.0.0.1/32 -j ACCEPT
-A OUTPUT -s x.x.x.0/24 -j ACCEPT
-A OUTPUT -s 127.0.0.1/32 -j ACCEPT
-A OUTPUT -s 127.0.0.1/32 -j ACCEPT
COMMIT
# Completed on Mon Jun 24 19:40:24 2019
```

!!! info
    If something goes wrong during the iptables configuration, you can always run ``` sudo iptables -F ```. Thats it! Your iptables are reset to default settings i.e. accept all! (do not forget to save the iptables configuration after you reset it, or it will be lost)



