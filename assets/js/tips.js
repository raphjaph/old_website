async function makeRequest(rune, method, params) {
  const LNSocket = await lnsocket_init()
  const ln = LNSocket()

  ln.genkey()
  await ln.connect_and_init("02b02f856f28cbe658133008b9dcb9ae2e6c18d27fbe5cd6644b6f13bcb42a269c", "wss://ln.8el.eu/10.13.13.2")

  // {} unwraps the promise
  const {result} = await ln.rpc({ rune, method, params })

  ln.disconnect()
  return result
};

function getInvoice(label, description) {
  const rune = "DG2_hQywgzXPdJlJoq64hdWhIRwCSvesoWLSZeZ945Q9OSZtZXRob2Q9aW52b2ljZQ==" 

  const params = {
    msatoshi: "any",
    label: label,
    description: description 
  }

  return makeRequest(rune, "invoice", params)
}

function waitInvoice(label) {
  const rune = "ro_I6rX06qrUVSgcZkAXPFwqtE2KOGGFJGsOnkayUQM9MTAmbWV0aG9kPXdhaXRpbnZvaWNl"

  return makeRequest(rune, "waitinvoice", { label })
}

function getInfo() {
  const rune = "<getinfo-rune>"
  
  return makeRequest(rune, "getinfo") 
}

async function onClickTipButton() {
  const tipButton = document.querySelector(".button")
  tipButton.style.color = "transparent"
  tipButton.classList.add("button--loading")
  
  const label = `tips/${new Date().getTime()}`

  // TODO: get small note from user
  description = prompt("Leave a note!", "")

  const invoice = await getInvoice(label, description)
 
  const link = "lightning:" + invoice.bolt11
  qr = new QRCode("qrcode", {
    text: link,
    width: 256,
    height: 256,
    colorDark : "#000000",
    colorLight : "#ffffff",
    correctLevel : QRCode.CorrectLevel.L
  })

  invoiceLink = document.querySelector("#invoice-link")
  invoiceLink.href = link

  tipButton.style.display = "none"

  const paid = await waitInvoice(label)
  if (paid.status === "paid") {
    qrcode = document.getElementById("qrcode")
    qrcode.innerHTML = "<br/> 🤑 <br/><br/> Thanks! <br/><br/>"
    qrcode.style.fontSize = "50px";
  }

}

//    function getInvoiceThroughAPI() {
//            fetch("https://raph.8el.eu/api/getinvoice")
//                .then( r => r.text())
//                .then( invoice => { 
//                        console.log(invoice)
//
//                        link = "lightning:" + invoice
//
//                        qr = new QRCode("qrcode", {
//                                text: link,
//                                width: 256,
//                                height: 256,
//                                colorDark : "#000000",
//                                colorLight : "#ffffff",
//                                correctLevel : QRCode.CorrectLevel.L
//                            })
//
//                        invoiceLink = document.querySelector("#invoice-link")
//                        invoiceLink.href = link
//
//                        tipButton = document.getElementById("tip-button")
//                        tipButton.style.display = "none"
//                    })
//        };
