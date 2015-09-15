function doVote(factId, up) {

	voteRequest = {FactId: factId, Up: up===true};

	var request = JSON.stringify(voteRequest);
	var xmlhttp = new XMLHttpRequest();
	var url = "/api/vote"

	xmlhttp.onreadystatechange = function() {
		if (xmlhttp.readyState == 4 && xmlhttp.status == 200) {
			var response = JSON.parse(xmlhttp.responseText);
			processResponse(response);
		} else if(xmlhttp.readyState == 4 && xmlhttp.status != 200) {
			alert("Error: " + xmlhttp.responseText);
		}
	}

	xmlhttp.open("POST", url, true);
	xmlhttp.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
	xmlhttp.send(request);
}

function processResponse(response) {
	if(response.Response == "ok") {
		//vote was OK!
		scoreFieldId = "fact-"+response.FactId+"-score";
		yourScoreFieldId = "fact-"+response.FactId+"-yourscore";
		yourAccountBankFieldId = "account-votebank";

		document.getElementById(scoreFieldId).innerHTML = "" + response.NewScore.Ups + "/" + response.NewScore.Downs;
		document.getElementById(yourScoreFieldId).innerHTML = "" + response.NewScore.AccountVote;
		document.getElementById(yourAccountBankFieldId).innerHTML = "" + response.NewVoteBank;

	} else {
		alert(response.Response);
	}
}

function doModerate(factId, enable) {

	voteRequest = {FactId: factId, Enable: enable===true};

	var request = JSON.stringify(voteRequest);
	var xmlhttp = new XMLHttpRequest();
	var url = "/api/moderate"

	xmlhttp.onreadystatechange = function() {
		if (xmlhttp.readyState == 4 && xmlhttp.status == 200) {
			var response = JSON.parse(xmlhttp.responseText);
			processModerateResponse(response);
		} else if(xmlhttp.readyState == 4 && xmlhttp.status != 200) {
			alert("Error: " + xmlhttp.responseText);
		}
	}

	xmlhttp.open("POST", url, true);
	xmlhttp.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
	xmlhttp.send(request);
}

function processModerateResponse(response) {
	if(response.Response == "ok") {
		//moderate was OK!

		console.log("Got response: " + JSON.stringify(response));

		moderateFieldId = "fact-"+response.FactId+"-moderate";
		
		moderateFieldLink = "fact-"+response.FactId+"-moderatelink";

		if(response.NewAwaitModeration == false) {
			document.getElementById(moderateFieldLink).innerHTML = "Moderator - Disable";
			document.getElementById(moderateFieldLink).setAttribute("onclick", "doModerate("+response.FactId+", true)");
			
			document.getElementById(moderateFieldId).style.display = "none";
		} else {
			document.getElementById(moderateFieldLink).innerHTML = "Moderator - Approve";
			document.getElementById(moderateFieldLink).setAttribute("onclick", "doModerate("+response.FactId+", false)");
			
			document.getElementById(moderateFieldId).style.display = "inline";
		}
		

	} else {
		alert(response.Response);
	}
}

var nextReferenceId = 2;

function addReference() {
	refsContainer = document.getElementById("references")

	refsContainer.innerHTML += " "+
"<div id='reference"+nextReferenceId+"'> "+                  
"	<h2 class='content-subhead'>Reference "+(nextReferenceId+1)+"</h2> "+
"    <div class='pure-control-group'> "+
"    	<label for='References."+nextReferenceId+".Url'>URL</label> "+
"    	<input class='pure-input-2-3' id='References."+nextReferenceId+".Url' name='References."+nextReferenceId+".Url' type='text' placeholder='http://example.com' required autocomplete='off'> "+
"    </div> "+
" "+
"     <div class='pure-control-group'> "+
"    	<label for='References."+nextReferenceId+".Publisher'>Author/Publisher</label> "+
"    	<input class='pure-input-2-3' id='References."+nextReferenceId+".Publisher' name='References."+nextReferenceId+".Publisher' type='text' placeholder='Example Media Corp' required autocomplete='off'> "+
"    </div> "+
" "+
"     <div class='pure-control-group'> "+
"    	<label for='References."+nextReferenceId+".Title'>Page Title</label> "+
"    	<input class='pure-input-2-3' id='References."+nextReferenceId+".Title' name='References."+nextReferenceId+".Title' type='text' placeholder='An example webpage' required autocomplete='off'> "+
"</div> "+
"	";
	nextReferenceId++;
}

function removeReference() {
	nextReferenceId--;
	
	try{
	lastReference = document.getElementById("reference"+nextReferenceId);

	lastReference.parentNode.removeChild(lastReference);
	}catch(e){
		nextReferenceId++;
	}
}

