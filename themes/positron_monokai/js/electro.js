var App = App || {}; // Create namespace

(function () {
    "use strict";

    App.state = {
        allMenuSpans: null,
        changeBars: [],
    };
    App.SEARCH_RESULT_SNIPPET_MAX_LEN = 200;

    App.main = () => {
        // Show version
        console.log('Built with Electro ' + App.globalConfig.electroVersion);

        const pageId = App.getUrlValue('pageId');
        const headingId = App.getUrlValue('headingId');

        // -----------------------
        // menu-tree: Find all items (spans) in all menu-tree(s)
        // -----------------------
        const menu_trees = document.getElementsByClassName("menu-tree");
        App.state.allMenuSpans = [];
        for (let menu_tree of menu_trees) {
            const menu_spans = menu_tree.getElementsByTagName("span");
            for (let span of menu_spans) {
                App.state.allMenuSpans.push(span);
            }
        }

        // -----------------------
        // menu-tree: Set item (span) click handler
        // -----------------------
        for (let span of App.state.allMenuSpans) {
            span.addEventListener("click", function () {
                App.onClickMenuItem(this);
            });
        }

        // -----------------------
        // menu-tree: Select the current document (at page load)
        // -----------------------
        if(App.globalConfig.singleFile){
            if (pageId){
                App.selectMenuItem(pageId, headingId);
            }
        }
        else {
            // Static site
            App.selectMenuItem(App.globalConfig.currentDocumentName, null);
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
        // console.log(searchBox);
        searchBox.addEventListener(
            "keydown",
            function (event) {
                // Press: Enter
                if (event.keyCode == 13) {
                    event.preventDefault();
                    if (App.globalConfig.singleFile){
                        const searchText = document.getElementById("search-text").value;
                        App.doSearch(searchText);
                    }
                    else{
                        App.onSearch();
                    }
                }
            },
            false
        );

        // -----------------------
        // Home button
        // -----------------------
        if(App.globalConfig.singleFile){
            // There are 2 versions of the home button. The default is an anchor tag
            // to index.html, for static web sites.  For single file docs we need to hide that
            // one and enable the one which contains a pageId link parameter (to the doc_info page).
            document.getElementById("home-button-static").style.display = 'none';
            document.getElementById("home-button-single-file").style.display = 'inline';
        }

        // -----------------------
        // Get search data
        // -----------------------
        const config = App.globalConfig;
        config.searchIndex = lunr(function () {
            this.field("title");
            this.field("text");
            this.field("heading");
            this.ref("location");

            for (var i = 0; i < App.searchData.docs.length; i++) {
                var doc = App.searchData.docs[i];
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
        
        // Navigate to search result location (if requested in url)
        if(pageId){
            console.log('Going to location: pageId:' + pageId + ' headingId:' + headingId);
            App.showPage(pageId);
            if(headingId){
                App.scrollToHash(headingId);
            }
        }

        const hideWatermark = App.getUrlValue('nowm');
        if (hideWatermark == 'true'){
            console.log('Hiding watermark because nowm set in URL.');
            document.getElementById("watermark-text").innerHTML = "";
        }

        // -----------------------
        //  Set handler
        // -----------------------
        const main_element = document.getElementsByClassName("main-container")[0];
        main_element.addEventListener('touchmove', App.onMainTouchMove, false);

        window.onload = function() {
            App.initializeChangeBars();
            App.updateChangeBars();
        };
    };

    // Highlight the sidebar menu item associated with documentName and (if non-null) headingName.
    App.selectMenuItem = (documentName, headingName) => {
        var spanDocument = null;
        var spanHeading = null;
        console.log('selectMenuItem(): ' + documentName + " :: " + headingName);
        for (let span of App.state.allMenuSpans) {
            if (
                (span.id == "menuitem_doc_" + documentName)
                && (span.classList.contains("level-0"))
            ){
                // Found document menu item
                spanDocument = span;
            }
            if (
                headingName
                && (span.id == "menuitem_doc_" + documentName)
                && (span.dataset.targetHeadingId == headingName)
            ){
                // Found heading menu item
                spanHeading = span;
            }
        }
        if (!spanDocument){
            console.log('   document ' + documentName + ' not found.');
            return;
        }
        if(!spanHeading){
            console.log('   heading ' + headingName + ' not found. Selecting document menu item.');
            spanDocument.classList.add("selected");
        } else{
            console.log('   heading ' + headingName + ' found. Selecting heading menu item.');
            // Unfold the ul associated with the document menu item (so that the heading menu
            // item will be visible).
            var list = spanDocument.parentElement.getElementsByTagName('ul')[0];
            list.classList.add("active");

            // Select the heading menu item
            spanHeading.classList.add("selected");
        }
    };

    App.onClickMenuItem = (self) => {
        // console.log("App.onClickMenuItem");
        // Unselect all items in all menu-tree(s)
        for (let old_span of App.state.allMenuSpans) {
            old_span.classList.remove("selected");
            old_span.classList.remove("navigating");
        }
        
        if (App.globalConfig.singleFile){
            App.onClickMenuSingleFile(self);
        }
        else{
            App.onClickMenuStaticSite(self);
        }
    };
    
    App.onClickMenuSingleFile = (self) => {
        // console.log("App.onClickMenuSingleFile");
        const pageId = self.dataset.documentName;
        const targetHeadingId = self.dataset.targetHeadingId;
        
        // -----------------------------
        // Make the target page visible
        // -----------------------------
        App.showPage(pageId);

        // ----------------------------
        // Scroll to the target heading (or to top if no HeadingId exists)
        // -----------------------------
        App.scrollToHash(targetHeadingId);

        // -----------------------------
        // Select the clicked menu item
        // -----------------------------
        self.classList.add("selected");

        // If in responsive narrow-screen, re-hide the menu.
        document.getElementById("sidebar-container").classList.remove("force-show");
    };

    App.onClickMenuStaticSite = (self) => {
        // console.log("App.onClickMenuStaticSite");
        // Navigate to link, if this span contains one
        var target_url = null;
        const anchors = self.getElementsByTagName("a");
        if (anchors) {
            for (let anchor of anchors) {
                target_url = anchor.href;
                // There should be only one anchor, but break anyway.
                break;
            }
        }
        if (target_url !== null){
            // if(App.globalConfig.currentDocumentName in 
            if (target_url.indexOf(App.globalConfig.currentDocumentName) > -1) {
                // We are navigating to the current page
                self.classList.add("selected");
            }
            else{
                // We are navigating away from the current page
                self.classList.add("navigating");
            }
            window.location.href = target_url;
        }
        else {
            self.classList.add("selected");
        }
    };

    App.showPage = (pageId) => {
        const pages = document.getElementsByClassName('content-page');
        for (const page of pages) {
            if(page.id == pageId){
                page.style.display = 'inline';
            } else {
                page.style.display = 'none';
            }
        }
    };

    App.scrollToHash = (hashName) => {
        if(!hashName){
            // Scroll to the top if no hashName was given
            var content_div = document.getElementsByClassName("main-container")[0];
            content_div.scrollTop = 0;
            return;
        }
        const target = "#" + hashName;
        if(location.hash == target){
            // Browser thinks we are already at the target, but we might have scrolled away
            // from it, so blank it and then set it back to force the document to scroll
            location.hash = "";
            location.hash = "#" + hashName;
        }
        else {
            location.hash = "#" + hashName;
        }
    };

    App.onMainTouchMove = (e) => {
        if(document.getElementById("sidebar-container").classList.contains("force-show")){
            // Don't allow the main window to be scrolled while the sidebar is 
            // deployed in "mobile" (narrow screen) mode.
            e.preventDefault();
        }
    };

    App.toggleSidebar = () => {
        console.log('toggleSidebar()');
        document.getElementById("sidebar-container").classList.toggle("force-show");
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
        return null;
    };

    App.onSearch = () => {
        // Executed on static sites when user enters search text and presses return.
        //
        // The search is encoded as a query parameter on the URL, and the search page
        // (search.html) is loaded. The actual searching and displaying of results will
        // happen on that page.
        console.log("onSearch()");
        var searchText = document.getElementById("search-text").value;
        searchText = searchText.trim();
        if(!searchText){
            console.log('Search text blank. Ignoring.');
            return;
        }
        console.log("searchText", searchText);
        window.location.href = 'search.html?query=' + searchText;
    };

    App.doSearch = (searchText) => {
        // console.log("doSearch()");
        searchText = searchText.trim();
        if(!searchText){
            console.log('Search text blank. Ignoring.');
            return;
        }
        
        const results = App.search(searchText);
        // console.log(results);

        var html = '<h1>Search Results</h1>\n';
        results.forEach(result => {
            // html += '<hr>';
            if (App.globalConfig.singleFile){
                // Single file compatible link
                const location = result.location.replace(".html", "");
                const pieces = location.split("#");
                const pageId = pieces[0];
                const headingId = pieces[1];
                if(result.heading){
                    // Show document heading only, as link
                    html += '<h4><a href="?pageId=' + pageId + '&headingId=' + headingId + '">' + result.heading + '</a></h4>';
                }
                else{
                    // Show document name only, as link
                    html += '<h3><a href="?pageId=' + pageId + '&headingId=' + headingId + '">' + result.title + '</a></h3>';
                }
            } else {
                // Static site compatible link
                html += '<h3><a href="' + result.location + '">' + result.title + '</a></h3>';
                if(result.heading){
                    html += '<h4>' + result.heading + '</h4>';
                }
            }

            // Highlight search hits
            const summary = App.highlight(searchText, result.summary);

            html += '<p>' + summary + '</p>';
        });
        if (results.length == 0){
            html += '(no results to show)';
        }
        if(!App.globalConfig.singleFile){
            // Static web page. Show search results in content area
            document.getElementById("content").innerHTML = html;
        } else {
            // Single file document. Show search results in search page and hide other pages.
            document.getElementById("search").innerHTML = html;

            // Make the search page visible
            App.showPage('search');
        }
    };

    App.highlight = (searchText, html) => {
        // Case insensitive highlight searchText in html.
        // Wraps hits with '<span class="highlight">' tag.
        const regex = new RegExp(searchText, "ig");
        const matches = html.matchAll(regex);
        const match_words = [];
        var match;
        for(match of matches){
            match_words.push(match[0]);
        }
        const unique = Array.from(new Set(match_words));
        for(match of unique){
            html = html.replace(
                match, '<span class="highlight">' + match + '</span>');
        }
        return(html);
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
            doc.summary = App.truncateSearchResult(doc.text, query);
            resultDocuments.push(doc);
        }
        return resultDocuments;
    };

    /**
     * Shorten the text of a search result to a "snippet" no longer than some max, 
     * while attempting to include the search term(s) within that snippet.
     * @param {string} text The search result text.
     * @param {string} query The search query.
     * @return {string} The snippet text.
     */
    App.truncateSearchResult = (text, query) => {
        const compareText = text.toLowerCase();
        query = query.toLocaleLowerCase();

        const searchWords = query.trim().split(/\s+/);
        var hitIndex = -1;
        for (const searchWord of searchWords){
            const i = compareText.indexOf(searchWord);
            if(i != -1){
                if(hitIndex == -1){
                    // First hit found
                    hitIndex = i;
                }
                else{
                    // Take the earliest of all hits found
                    hitIndex = Math.min(i, hitIndex);
                }
            }
        }
        if(hitIndex == -1){
            // If we were unable to find any of the search terms literally then just set
            // the hitIndex to 0, which will cause us to create a snippet beginning at the
            // start of the text.
            hitIndex = 0;
        }
        var size = text.length;
        var start = hitIndex - Math.floor(App.SEARCH_RESULT_SNIPPET_MAX_LEN/2);
        var end = hitIndex + Math.floor(App.SEARCH_RESULT_SNIPPET_MAX_LEN/2);
        var overflow;

        // Slide region left to not overhang end
        overflow = end - (size - 1);
        if (overflow > 0){
            end -= overflow;
            start -= overflow;
        }
        // Slide region right to not overhand start
        overflow = -start;
        if (overflow > 0){
            start += overflow;
            end += overflow;
        }
        // Shorten region if longer than text
        if (end > size-1){
            end = size-1;
        }
        // Extract snippet and add ellipses as necessary
        var snippet = (start > 0) ? "…" : "";
        snippet += text.substring(start, end);
        if (end < size - 1){
            snippet += "…";
        }
        // console.log({query: query, text:text, start:start, end:end, snippet:snippet});
        // console.log(text);
        return snippet;
    };

    App.initializeChangeBars = function() {
        console.log('App.initializeChangeBars()');
        App.state.changeBars = [];
        var elements = document.getElementsByClassName('content-left-gutter');
        const elGutter = elements[0];
        elements = document.getElementsByClassName('anchor-change-bar');
        // console.log(elements);
        // let startY = 0;
        let elAnchorStart = undefined;
        Array.from(elements).forEach(function(el){
            console.log(el);
            if(el.classList.contains('start')){
                elAnchorStart = el;
                // startY = elAnchorStart.offsetTop;
                // console.log('Found startY', startY);
            } else if (el.classList.contains('end') && elAnchorStart !== undefined){
                let elAnchorEnd = el;
                // let endY = elAnchorEnd.offsetTop;
                // console.log('Found endY', endY);

                const elChangeBar = document.createElement("div");
                elChangeBar.setAttribute('class', 'change-bar');
                // elChangeBar.style.top = startY + 'px';
                // elChangeBar.style.height = (endY - startY) + 'px';
                elGutter.appendChild(elChangeBar);



                App.state.changeBars.push({
                    elAnchorStart: elAnchorStart,
                    elAnchorEnd: elAnchorEnd,
                    elChangeBar: elChangeBar
                });
                
            }
        });
        // $('.anchor-change-bar').each(function (x){
        //     console.log(this);
        //     console.log("X", x);
        // });
    };

    App.updateChangeBars = function() {
        console.log('App.updateChangeBars()');
        App.state.changeBars.forEach(function(cb){
            const startY = cb.elAnchorStart.offsetTop;
            const endY = cb.elAnchorEnd.offsetTop;
            cb.elChangeBar.style.top = startY + 'px';
            cb.elChangeBar.style.height = (endY - startY) + 'px';
            console.log(cb);
        });
    };
})(); // "use strict" wrapper
