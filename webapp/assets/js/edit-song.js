import {Transposer} from 'https://cdn.skypack.dev/chord-transposer';

window.addEventListener('DOMContentLoaded', (e) => {

    Telegram.WebApp.expand()

    let form = document.getElementById('form')
    let name = document.getElementById('name')
    autosize(name)
    let key = document.getElementById('key')
    let bpm = document.getElementById('bpm')
    let time = document.getElementById('time')
    let tags = document.getElementById("tags")
    let newTag = document.getElementById("new-tag")

    tags.onchange = function (e) {
        let isCustom = Array.from(this.selectedOptions).find(value => value.value === "custom");

        if (isCustom) {
            newTag.hidden = false;
            newTag.focus();
        } else {
            newTag.hidden = true;
        }
    }

    // new TomSelect('#tags', {
    //     plugins: ['remove_button', 'input_autogrow'],
    //     create: true,
    //     // createFilter: /^\p{Lu}\p{Ll}+( \p{L}+)*$/gmu,
    //     placeholder: "Добавить тег"
    // });

    Telegram.WebApp.ready()

    if (action === "edit") {
        if (Array.from(key.options).find(o => o.value === song.pdf.key)) {
            key.value = song.pdf.key;
        } else {
            try {
                new Transposer(song.pdf.key).getKey()
                const o = new Option(song.pdf.key, song.pdf.key)
                console.log(o)
                key.add(o, 0)
                key.value = song.pdf.key
            } catch (err) {
                key.value = "?"
            }
        }
    }

    Array.from(form.elements).forEach((element) => {
        if (element.tagName === "SELECT" && element.multiple) {
            element.initValue = Array.from(element.selectedOptions).map(({value}) => value)
        } else {
            element.initValue = element.value
        }
    });

    form.addEventListener("submit", (e) => e.preventDefault())
    form.addEventListener('input', (e) => {

        let hide = []
        Array.from(form.elements).forEach((element) => {
            if (element.tagName === "SELECT" && element.multiple) {
                let opts = Array.from(element.selectedOptions).map(({value}) => value)
                hide.push(JSON.stringify(opts) === JSON.stringify(element.initValue))
            } else {
                // console.log(element)
                // console.log("init " + element.initValue)
                // console.log("curr " + element.value)
                hide.push(element.initValue === element.value)
            }
        });

        if (!hide.includes(false)) {
            Telegram.WebApp.MainButton.hide()
            console.log("hide")
        } else {
            Telegram.WebApp.MainButton.show()
            console.log("show")
        }
    })

    if (action === "create") {
        createSong()
    } else {
        editSong()
    }

    function createSong() {
        let lyrics = document.getElementById("lyrics")
        autosize(lyrics)

        Telegram.WebApp.MainButton.setText(createText)

        Telegram.WebApp.MainButton.onClick(function () {

            if (form.checkValidity() === false) {
                form.reportValidity()
                return
            }
            Telegram.WebApp.MainButton.showProgress()

            let t = Array.from(tags.selectedOptions)
                .map(({value}) => {
                    if (value === "custom") {
                        return newTag.value
                    }
                    return value;
                })
                .filter(value => value)

            let data = JSON.stringify({
                "name": name.value,
                "key": key.value,
                "bpm": bpm.value,
                "time": time.value,
                "tags": t,
                "lyrics": lyrics.value
            })


            Telegram.WebApp.sendData(data)
        })
    }

    function editSong() {
        Telegram.WebApp.MainButton.setText(saveText)

        let lyricsDiv = document.getElementById("lyrics")
        let initLyricsDiv = lyricsDiv.cloneNode(true)

        let transposeSection = document.getElementById("transpose-section")
        key.onchange = (e) => {

            if (e.target.value !== song.pdf.key) {
                transposeSection.classList.remove("visually-hidden")
                transposeSection.disabled = false
            } else {
                transposeSection.classList.add("visually-hidden")
                transposeSection.disabled = true
            }

            if (e.target.value === song.pdf.key) {
                lyricsDiv.innerHTML = initLyricsDiv.innerHTML
                return
            }

            let originalKey
            try {
                originalKey = new Transposer(song.pdf.key).getKey()
            } catch (err) {
                console.log(err)
            }

            let newKey
            try {
                newKey = new Transposer(e.target.value).getKey().majorKey
            } catch (err) {
                newKey = e.target.value
            }

            let clone = initLyricsDiv.cloneNode(true)
            let walker = document.createTreeWalker(
                clone,
                NodeFilter.SHOW_TEXT,
                null
            )

            while (walker.nextNode()) {
                if (originalKey) {
                    walker.currentNode.nodeValue = Transposer
                        .transpose(walker.currentNode.nodeValue)
                        .fromKey(originalKey)
                        .toKey(newKey).toString()
                } else {
                    walker.currentNode.nodeValue = Transposer
                        .transpose(walker.currentNode.nodeValue)
                        .toKey(newKey).toString()
                }
            }

            lyricsDiv.innerHTML = clone.innerHTML
        }

        Telegram.WebApp.MainButton.onClick(async function () {

                let transposeSection = document.getElementById("transpose-section")

                if (form.checkValidity() === false) {
                    form.reportValidity()
                    return
                }

                Telegram.WebApp.MainButton.showProgress()

                let t = Array.from(tags.selectedOptions)
                    .map(({value}) => {
                        if (value === "custom") {
                            return newTag.value
                        }
                        return value;
                    })
                    .filter(value => value)

                console.log(t)

                let data = JSON.stringify({
                    "name": name.value,
                    "key": key.value,
                    "transposeSection": transposeSection.value,
                    "bpm": bpm.value,
                    "time": time.value,
                    "tags": t
                })

                console.log(newTag.value)

                await fetch(`/web-app/songs/${song.id}/edit/confirm?queryId=${Telegram.WebApp.initDataUnsafe.query_id}&messageId=${messageId}&chatId=${chatId}&userId=${userId}`, {
                    method: "POST",
                    headers: {'Content-Type': 'application/json'},
                    body: data,
                })

                Telegram.WebApp.close()
            }
        )

    }
})

// console.log (lyricsContainer.offsetWidth - lyricsDiv.offsetWidth)
//  if ((lyricsContainer.offsetWidth - lyricsDiv.offsetWidth) < 0 || (lyricsContainer.offsetWidth - lyricsDiv.offsetWidth) > 0) {
//      let fontSize = 24
//      for (let i = 0; i < 20; i++) {
//          console.log(lyricsContainer.offsetWidth - lyricsDiv.offsetWidth)
//          if ((lyricsContainer.offsetWidth - lyricsDiv.offsetWidth) < 0) {
//              break
//          }
//          fontSize -= 1
//          lyricsDiv.style.fontSize = fontSize + "px !important"
//      }
//  }

// lyricsDiv.style.fontSize = "10px";
// let fontSize = 10
// console.log("lyricsDiv.offsetWidth " + lyricsDiv.offsetWidth)
// console.log(" lyricsContainer.offsetWidth " + lyricsContainer.offsetWidth)
// while (lyricsDiv.offsetWidth < lyricsContainer.offsetWidth) {
//     console.log("lyricsDiv.offsetWidth " + lyricsDiv.offsetWidth)
//     console.log(" lyricsContainer.offsetWidth " + lyricsContainer.offsetWidth)
//     lyrics.style.fontSize = fontSize + "px";
//     fontSize += 1;
// }
// while (lyricsDiv.offsetWidth >= lyricsContainer.offsetWidth) {
//     console.log("lyricsDiv.offsetWidth " + lyricsDiv.offsetWidth)
//     console.log(" lyricsContainer.offsetWidth " + lyricsContainer.offsetWidth)
//     fontSize -= 1;
//     lyrics.style.fontSize = fontSize + "px";
// }
