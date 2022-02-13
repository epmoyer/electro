var App = App || {}; // Create namespace

(function () {
    "use strict";

    App.main = () => {
        console.log("main()");

        // -----------------------
        // menu-tree: Find all items (spans) in all menu-tree(s)
        // -----------------------
        const menu_trees = document.getElementsByClassName("menu-tree");
        const all_menu_spans = [];
        for (let menu_tree of menu_trees) {
            const menu_spans = menu_tree.getElementsByTagName("span");
            // console.log(menu_spans);
            for (let span of menu_spans) {
                all_menu_spans.push(span);
            }
        }

        // -----------------------
        // menu-tree: Set item (span) click handler
        // -----------------------
        for (let span of all_menu_spans) {
            span.addEventListener("click", function () {
                // Unselect all items in all menu-tree(s)
                for (let old_span of all_menu_spans) {
                    old_span.classList.remove("selected");
                }
                // console.log("Clicked a span");
                this.classList.toggle("selected");
                // Navigate to link, if this span contains one
                const anchors = this.getElementsByTagName("a");
                if (anchors) {
                    for (let anchor of anchors) {
                        window.location.href = anchor.href;
                        // There should be only one anchor, but break anyway.
                        break;
                    }
                }
            });
        }

        // -----------------------
        // menu-tree: Select the current document (at page load)
        // -----------------------
        for (let span of all_menu_spans) {
            if (span.id == "menuitem_doc_" + App.globalConfig.currentDocumentName) {
                span.classList.add("selected");
            }
        }

        // -----------------------
        // menu-tree: Set caret (child folding) click handler
        // -----------------------
        var toggler = document.getElementsByClassName("caret");
        var i;

        for (i = 0; i < toggler.length; i++) {
            toggler[i].addEventListener("click", function (e) {
                this.closest("li").querySelector(".nested").classList.toggle("active");
                this.classList.toggle("caret-down");
                e.stopPropagation();
            });
        }

        // -----------------------
        // Search handler
        // -----------------------
        var searchBox = document.getElementById("search-text");
        console.log(searchBox);
        searchBox.addEventListener(
            "keydown",
            function (event) {
                if (event.keyCode == 13) {
                    event.preventDefault();
                    App.onSearch();
                }
            },
            false
        );

        // -----------------------
        // Get search data
        // -----------------------
        fetch("search/search_index.json")
            .then((response) => {
                return response.json();
            })
            .then((data) => {
                const config = App.globalConfig;
                console.log("Fetched search index");
                // console.log(data);
                config.searchData = data;
                config.searchIndex = lunr(function () {
                    this.field("title");
                    this.field("text");
                    this.field("heading");
                    this.ref("location");

                    for (var i = 0; i < data.docs.length; i++) {
                        var doc = data.docs[i];
                        this.add(doc);
                        config.documents[doc.location] = doc;
                    }
                });

                // Execute search query (if requested in url)
                const searchText = App.getUrlValue('query');
                if(searchText){
                    document.getElementById("search-text").value = searchText;
                    console.log('query', searchText);
                    App.doSearch(searchText);
                }
            });

    };

    App.getUrlValue  = (VarSearch) => {
        var SearchString = window.location.search.substring(1);
        var VariableArray = SearchString.split('&');
        for(var i = 0; i < VariableArray.length; i++){
            var KeyValuePair = VariableArray[i].split('=');
            if(KeyValuePair[0] == VarSearch){
                return decodeURI(KeyValuePair[1]);
            }
        }
    };

    App.onSearch = () => {
        console.log("onSearch()");
        var searchText = document.getElementById("search-text").value;
        console.log("searchText", searchText);
        // App.doSearch(searchText);
        window.location.href = 'search.html?query=' + searchText;
    };

    App.doSearch = (searchText) => {
        console.log("doSearch()");
        
        const results = App.search(searchText);
        console.log(results);

        var html = '<h1>Search Results</h1>\n';
        results.forEach(result => {
            html += '<hr>';
            html += '<h3><a href="/' + result.location + '">' + result.title + '</a></h3>';
            if(result.heading){
                html += '<h4>' + result.heading + '</h4>';
            }
            html += '<p>' + result.summary + '</p>'
        });
        if (results.length == 0){
            html += '(no results to show)';
        }
        document.getElementById("content").innerHTML = html;
    };

    App.search = (query) => {
        const config = App.globalConfig;
        if (config.searchIndex === null) {
            console.error("Assets for search still loading");
            return;
        }

        var resultDocuments = [];
        var results = config.searchIndex.search(query);
        for (var i = 0; i < results.length; i++) {
            var result = results[i];
            var doc = config.documents[result.ref];
            doc.summary = doc.text.substring(0, 200);
            resultDocuments.push(doc);
        }
        return resultDocuments;
    };
})(); // "use strict" wrapper
