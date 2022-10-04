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
            input: instantSearch.querySelector(".instant-search__input"),
            icon: instantSearch.querySelector(".instant-search__icon"),
            resultsContainer: document.createElement("div")
        };

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

            let query = this.elements.input.value;

            let setlist = query.split("\n")
            if (setlist.length > 1) {
                query = setlist[0]
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
            this.elements.resultsContainer.classList.add(
                "instant-search__results-container--visible"
            );
        });

        this.elements.input.addEventListener("blur", () => {
            this.elements.resultsContainer.classList.remove(
                "instant-search__results-container--visible"
            );
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
    }

    /**
     * Creates the HTML to represents a single result in the list of results.
     *
     * @param {Object} result An instant search result
     * @returns {HTMLAnchorElement}
     */
    createResultElement(result) {
        const anchorElement = document.createElement("a");

        anchorElement.classList.add("instant-search__result");
        anchorElement.insertAdjacentHTML(
            "afterbegin",
            this.options.templateFunction(result)
        );

        anchorElement.addEventListener("click", () => {
            this.populateResults([]);

            let setlist = this.elements.input.value.split("\n");
            setlist.shift()

            if (setlist.length > 0) {
                this.elements.input.value = setlist.join("\n");
                this.elements.input.dispatchEvent(new Event("input"));
            } else {
                this.elements.input.value = "";
            }
        })
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

export default InstantSearch;