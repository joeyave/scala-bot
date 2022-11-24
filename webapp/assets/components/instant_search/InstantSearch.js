class InstantSearch {
    /**
     * Initialises the instant search bar. Retrieves and creates elements.
     *
     * @param {HTMLElement} instantSearch The container element for the instant search
     * @param {InstantSearchOptions} options A list of options for configuration
     */
    constructor(instantSearch, options) {
        this.options = options;
        this.elements = {
            main: instantSearch,
            inputContainer: instantSearch.querySelector(".instant-search__input-container"),
            input: instantSearch.querySelector(".instant-search__input"),
            icon: instantSearch.querySelector(".instant-search__icon"),
            resultsContainer: document.createElement("div")
        };

        this.elements.resultsContainer.id = "instant-search-results-container"
        this.elements.resultsContainer.classList.add(
            "instant-search__results-container"
        );
        this.elements.main.appendChild(this.elements.resultsContainer);

        this.addListeners();
    }

    /**
     * Adds event listeners for elements of the instant search.
     */
    addListeners() {
        let delay;

        this.elements.input.addEventListener("input", () => {
            clearTimeout(delay);

            let query = this.elements.input.value.replaceAll(/\(.*?\)|[1-9.()_]*/gi, "");

            let setlist = query.split("\n")
            if (setlist.length > 1) {
                query = setlist[0]
                console.log(query)
            }

            delay = setTimeout(() => {
                if (query.length < 2) {
                    this.populateResults([]);
                    return;
                }

                this.performSearch(query).then((results) => {
                    this.populateResults(results);
                });
            }, 500);
        });

        this.elements.input.addEventListener("focus", () => {
            if (this.elements.resultsContainer.childNodes.length > 0) {
                this.elements.resultsContainer.classList.add(
                    "instant-search__results-container--visible"
                );
                this.elements.inputContainer.classList.add("instant-search__input-container--with-results");
            }
        });

        // let isResultSelected = false;
        // this.elements.resultsContainer.addEventListener("mouseover", () => {
        //     console.log("true")
        //     isResultSelected = true;
        // })
        // this.elements.resultsContainer.addEventListener("mouseout", () => {
        //     console.log("false")
        //     isResultSelected = false;
        // })
        // this.elements.resultsContainer.addEventListener("touchstart", () => {
        //     console.log("true")
        //     isResultSelected = true;
        // })
        // this.elements.resultsContainer.addEventListener("touchend", () => {
        //     console.log("false")
        //     isResultSelected = false;
        // })
        //
        // this.elements.input.addEventListener("blur", () => {
        //     if (isResultSelected) {
        //         return
        //     }
        //     this.elements.resultsContainer.classList.remove(
        //         "instant-search__results-container--visible"
        //     );
        //     this.elements.inputContainer.classList.remove("instant-search__input-container--with-results");
        // });

        document.addEventListener("mousedown", (e) => {

            if (this.elements.main === e.target || this.elements.main.contains(e.target)) {
                return;
            }
            // if (e.target.id === this.elements.main.id ||
            //     e.target.id === this.elements.input ||
            //     e.target.id === this.elements.resultsContainer.id ||
            // e.target.className.contains("instant-search__title")) {
            //     return
            // }
            this.elements.resultsContainer.classList.remove(
                "instant-search__results-container--visible"
            );
            this.elements.inputContainer.classList.remove("instant-search__input-container--with-results");
        });
    }

    /**
     * Updates the HTML to display each result under the search bar.
     *
     * @param {Object[]} results
     */
    populateResults(results) {
        // Clear all existing results
        while (this.elements.resultsContainer.firstChild) {
            this.elements.resultsContainer.removeChild(
                this.elements.resultsContainer.firstChild
            );
        }

        // Update list of results under the search bar
        for (const result of results) {
            this.elements.resultsContainer.appendChild(
                this.createResultElement(result)
            );
        }


        if (results.length > 0) {
            this.elements.inputContainer.classList.add("instant-search__input-container--with-results");
            this.elements.resultsContainer.classList.add("instant-search__results-container--visible");
            Telegram.WebApp.HapticFeedback.impactOccurred("light");
        } else {
            this.elements.inputContainer.classList.remove("instant-search__input-container--with-results");
            this.elements.resultsContainer.classList.remove("instant-search__results-container--visible");
        }
    }

    /**
     * Creates the HTML to represents a single result in the list of results.
     *
     * @param {Object} result An instant search result
     * @returns {HTMLAnchorElement}
     */
    createResultElement(result) {
        const anchorElement = document.createElement("div");

        anchorElement.classList.add("instant-search__result");
        anchorElement.insertAdjacentHTML(
            "afterbegin",
            this.options.templateFunction(result)
        );

        // anchorElement.addEventListener("click", () => {
        // })

        anchorElement.addEventListener(...this.options.resultEventListener(result, this))

        // // If provided, add a link for the result
        // if ("href" in result) {
        //     anchorElement.setAttribute("href", result.href);
        // }

        return anchorElement;
    }

    /**
     * Makes a request at the search URL and retrieves results.
     *
     * @param {string} query Search query
     * @returns {Promise<Object[]>}
     */
    performSearch(query) {
        const url = new URL(this.options.searchUrl.toString());

        url.searchParams.set(this.options.queryParam, query);

        this.setLoading(true);

        return fetch(url, {
            method: "get"
        })
            .then((response) => {
                if (response.status !== 200) {
                    throw new Error("Something went wrong with the search!");
                }

                return response.json();
            })
            .then((responseData) => {

                return this.options.responseParser(responseData);
            })
            .catch((error) => {
                console.error(error);

                return [];
            })
            .finally((results) => {
                this.setLoading(false);

                return results;
            });
    }

    /**
     * Shows or hides the loading indicator for the search bar.
     *
     * @param {boolean} b True will show the loading indicator, false will not
     */
    setLoading(b) {
        this.elements.icon.classList.toggle("loader", b);
    }
}

function childOf(node, ancestor) {
    var child = node;
    while (child !== null) {
        if (child === ancestor) return true;
        child = child.parentNode;
    }
    return false;
}

export default InstantSearch;