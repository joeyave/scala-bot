import InstantSearch from "../components/instant_search/InstantSearch.js";

window.addEventListener('DOMContentLoaded', (e) => {

    Telegram.WebApp.expand()

    let form = document.getElementById('form')
    let name = document.getElementById("name")
    let date = document.getElementById("date")
    let search = document.getElementById("search")
    let songs = document.getElementById("songs")
    let notes = document.getElementById("notes")
    let searchInput = document.getElementById("song-search-input")
    autosize(notes)
    autosize(searchInput)

    let sortable = new Sortable(songs, {
        group: "songs",
        // forceFallback: true, // todo: set true on pc.
        delay: 150,
        delayOnTouchOnly: true,
        animation: 100,
        // Element is chosen
        onChoose: function (/**Event*/evt) {
            Telegram.WebApp.HapticFeedback.impactOccurred("light")
        },
        onMove: function (evt) {
            Telegram.WebApp.HapticFeedback.selectionChanged()
        },
        onUpdate: function (/**Event*/e) {
            let hide = []
            Array.from(form.elements).forEach((element) => {
                hide.push(element.initValue === element.value)
            });

            if (!hide.includes(false) && sortableInit === JSON.stringify(sortable.toArray())) {
                Telegram.WebApp.MainButton.hide()
                Telegram.WebApp.disableClosingConfirmation()
                console.log("hide")
            } else {
                Telegram.WebApp.MainButton.show()
                Telegram.WebApp.enableClosingConfirmation()
                console.log("show")
            }
        },

        filter: ".song-remove",
        onFilter: function (e) {

            // if (Sortable.utils.is(e.target, ".song-remove")) {
            Telegram.WebApp.HapticFeedback.selectionChanged();
            e.item.parentNode.removeChild(e.item);
            // sortableInit = JSON.stringify(sortable.toArray())
            Telegram.WebApp.MainButton.show()
            Telegram.WebApp.enableClosingConfirmation()
            // }
        },
    });

    let sortableInit = JSON.stringify(sortable.toArray());

    var reqUrl = `/api/drive-files/search?driveFolderId=${event.band.driveFolderId}`;
    if (event.band.archiveFolderId) {
        reqUrl += `&archiveFolderId=${event.band.archiveFolderId}`
    }
    new InstantSearch(search, {
        searchUrl: new URL(reqUrl, window.location.origin),
        queryParam: "q",
        responseParser: (responseData) => {
            return responseData.results;
        },
        templateFunction: (result) => {
            return `
            <div class="instant-search__title">${result.name}</div>
<!--            <p class="instant-search__paragraph">${result.occupation}</p>-->
        `;
        },
        resultEventListener: (result, search) => {
            return [
                "click", async () => {
                    console.log("click")

                    search.setLoading(true)
                    // Notiflix.Block.dots('.sortable-list');

                    document.getElementById("song-search-input").focus()

                    let resp = await fetch(`/api/songs/find-by-drive-file-id?driveFileId=${result.id}`, {
                        method: "get",
                        headers: {'Content-Type': 'application/json'},
                    })

                    let data = await resp.json()

                    console.log(JSON.stringify(data));

                    let spans = document.getElementById("songs").getElementsByTagName("span")

                    let exists = false
                    for (let i = 0; i < spans.length; i++) {
                        let songId = spans[i].getAttribute("data-song-id")
                        if (songId === data.song.id) {
                            exists = true
                            break
                        }
                    }

                    if (!exists) {
                        songs.insertAdjacentHTML("beforeend",
                            `<div class="item">
                            <span class="text" data-song-id=${data.song.id}>${result.name}</span>
                            <i class="fas fa-trash-alt song-remove"></i>
                        </div>`
                        );

                        Notiflix.Notify.success(songAddedText);
                        Telegram.WebApp.HapticFeedback.notificationOccurred("success");

                        // sortableInit = JSON.stringify(sortable.toArray())
                        Telegram.WebApp.MainButton.show()
                        Telegram.WebApp.enableClosingConfirmation()
                    } else {
                        Notiflix.Notify.warning(songExistsText);
                        Telegram.WebApp.HapticFeedback.notificationOccurred("warning");
                    }

                    search.elements.resultsContainer.classList.remove(
                        "instant-search__results-container--visible"
                    );
                    search.elements.inputContainer.classList.remove("instant-search__input-container--with-results");

                    search.populateResults([]);

                    let setlist = search.elements.input.value.split("\n");
                    setlist.shift()

                    if (setlist.length > 0) {
                        search.setLoading(true);
                        search.elements.input.value = setlist.join("\n");
                        search.elements.input.dispatchEvent(new Event("input"));
                    } else {
                        search.elements.input.value = "";
                    }

                    search.setLoading(false)
                    // Notiflix.Block.remove('.sortable-list');
                }
            ]
        }
    });

    Telegram.WebApp.ready()

    Array.from(form.elements).forEach((element) => {
        element.initValue = element.value
    });

    form.addEventListener("submit", (e) => e.preventDefault())
    form.addEventListener('input', (e) => {

        if (e.target.id === "song-search-input") {
            return
        }

        let hide = []
        Array.from(form.elements).forEach((element) => {
            hide.push(element.initValue === element.value)
        });

        if (!hide.includes(false) && sortableInit === JSON.stringify(sortable.toArray())) {
            Telegram.WebApp.MainButton.hide()
            Telegram.WebApp.disableClosingConfirmation()
            console.log("hide")
        } else {
            Telegram.WebApp.MainButton.show()
            Telegram.WebApp.enableClosingConfirmation()
            console.log("show")
        }
    })

    // document.addEventListener("click", (e) => {
    //     if (e.target.id === "delete-song-icon") {
    //         e.target.parentElement.remove()
    //         // Telegram.WebApp.MainButton.show()
    //     }
    // })

    if (action === "create") {
        createEvent()
    } else {
        editEvent(event)
    }


    function editEvent(event) {
        Telegram.WebApp.MainButton.setText(saveText)

        Telegram.WebApp.MainButton.onClick(async function () {

            if (form.checkValidity() === false) {
                form.reportValidity()
                return
            }
            Telegram.WebApp.MainButton.showProgress()

            let songIds = []
            let items = songs.getElementsByTagName("span")
            for (let i = 0; i < items.length; i++) {
                let songId = items[i].getAttribute("data-song-id")
                songIds.push(songId)
            }

            let data = JSON.stringify({
                "time": date.valueAsDate,
                "name": name.value,
                "bandId": event.bandId,
                "songIds": songIds,
                "notes": notes.value
            })

            await fetch(`/web-app/events/${event.id}/edit/confirm?queryId=${Telegram.WebApp.initDataUnsafe.query_id}&messageId=${messageId}&chatId=${chatId}&userId=${userId}&lang=${lang}`, {
                method: "POST",
                headers: {'Content-Type': 'application/json'},
                body: data,
            })

            Telegram.WebApp.close()
        })
    }

    function createEvent() {
        Telegram.WebApp.MainButton.setText(createText)

        Telegram.WebApp.MainButton.onClick(function () {

            if (form.checkValidity() === false) {
                form.reportValidity()
                return
            }
            Telegram.WebApp.MainButton.showProgress()

            let songIds = []
            let items = songs.getElementsByTagName("span")
            for (let i = 0; i < items.length; i++) {
                let songId = items[i].getAttribute("data-song-id")
                songIds.push(songId)
            }

            let data = JSON.stringify({
                "time": date.valueAsDate,
                "name": name.value,
                // "bandId": event.bandId,
                "songIds": songIds,
                "notes": notes.value
            })

            Telegram.WebApp.sendData(data)
        })
    }
});