class AudioController {
    constructor(audioElement) {
        this.element = audioElement;
    }

    setSource(url) {
        this.element.src = url;
    }

    clearSource() {
        this.element.removeAttribute("src");
    }

    load() {
        this.element.load();
    }

    play() {
        return this.element.play();
    }

    pause() {
        this.element.pause();
    }

    isPaused() {
        return this.element.paused;
    }

    getDuration() {
        return this.element.duration || 0;
    }

    getCurrentTime() {
        return this.element.currentTime || 0;
    }

    setCurrentTime(value) {
        this.element.currentTime = value;
    }

    setVolume(value) {
        this.element.volume = value;
    }

    getVolume() {
        return this.element.volume;
    }

    on(event, handler) {
        this.element.addEventListener(event, handler);
    }

    off(event, handler) {
        this.element.removeEventListener(event, handler);
    }
}

class MusicSpaApp {
    constructor(doc = document) {
        this.document = doc;
        this.root = doc.documentElement;
        this.apiBase = (this.root.dataset.apiBase || "/").replace(/\/?$/, "/");

        this.storageKeys = {
            playlist: "music-server-mobile:playlist",
            currentIndex: "music-server-mobile:current-index",
            volume: "music-server-mobile:volume",
            theme: "music-server-mobile:theme"
        };

        this.dom = {
            stats: doc.getElementById("library-stats"),
            status: doc.getElementById("playback-status"),
            nowPlayingTitle: doc.getElementById("now-playing-title"),
            nowPlayingArtist: doc.getElementById("now-playing-artist"),
            nowPlayingAlbum: doc.getElementById("now-playing-album"),
            cover: doc.querySelector(".cover"),
            playBtn: doc.getElementById("play-btn"),
            prevBtn: doc.getElementById("prev-btn"),
            nextBtn: doc.getElementById("next-btn"),
            progress: doc.getElementById("progress"),
            currentTime: doc.getElementById("current-time"),
            totalTime: doc.getElementById("total-time"),
            volume: doc.getElementById("volume-range"),
            searchForm: doc.getElementById("search-form"),
            searchInput: doc.getElementById("search-input"),
            searchResults: doc.getElementById("search-results"),
            playlistList: doc.getElementById("playlist-list"),
            playlistCount: doc.getElementById("playlist-count"),
            clearPlaylistBtn: doc.getElementById("clear-playlist-btn"),
            shuffleBtn: doc.getElementById("shuffle-btn"),
            audio: doc.getElementById("audio-player"),
            searchTemplate: doc.getElementById("search-result-template"),
            playlistTemplate: doc.getElementById("playlist-item-template"),
            themeToggle: doc.getElementById("theme-toggle")
        };

        this.audio = new AudioController(this.dom.audio);
        this.theme = "default";
        this.state = {
            playlist: [],
            currentIndex: -1,
            searchAbortController: null
        };
        this.debouncedSearch = this.debounce((value) => this.performSearch(value), 350);
    }

    init() {
        this.loadTheme();
        this.loadState();
        this.renderPlaylist();
        this.renderNowPlaying();
        this.bindEvents();
        void this.updateLibraryStats();
        this.loadCurrentTrack();
    }

    loadCurrentTrack() {
        if (this.state.currentIndex >= 0 && this.state.currentIndex < this.state.playlist.length) {
            const track = this.state.playlist[this.state.currentIndex];
            if (track && track.src) {
                this.loadTrack(track);
            }
        }
    }

    resolveUrl(path) {
        if (!path) {
            return "";
        }
        if (/^https?:\/\//i.test(path)) {
            return path;
        }
        const cleanBase = this.apiBase.endsWith("/") ? this.apiBase : `${this.apiBase}/`;
        return `${cleanBase}${path.replace(/^\//, "")}`;
    }

    formatTime(value) {
        if (Number.isNaN(value) || value == null) {
            return "0:00";
        }
        const seconds = Math.floor(value);
        const minutes = Math.floor(seconds / 60);
        const rest = seconds % 60;
        return `${minutes}:${rest < 10 ? "0" : ""}${rest}`;
    }

    saveState() {
        localStorage.setItem(this.storageKeys.playlist, JSON.stringify(this.state.playlist));
        localStorage.setItem(this.storageKeys.currentIndex, String(this.state.currentIndex));
    }

    loadState() {
        try {
            const rawPlaylist = localStorage.getItem(this.storageKeys.playlist) || "[]";
            const playlist = JSON.parse(rawPlaylist);
            if (Array.isArray(playlist)) {
                this.state.playlist = playlist;
            }
            const index = parseInt(localStorage.getItem(this.storageKeys.currentIndex) ?? "-1", 10);
            if (!Number.isNaN(index)) {
                this.state.currentIndex = index;
            }
            const volume = parseFloat(localStorage.getItem(this.storageKeys.volume) ?? "0.7");
            const normalizedVolume = Number.isNaN(volume) ? 0.7 : Math.min(Math.max(volume, 0), 1);
            this.audio.setVolume(normalizedVolume);
            this.dom.volume.value = String(this.audio.getVolume());
        } catch (err) {
            console.warn("Impossible de charger l'état du lecteur", err);
            this.state.playlist = [];
            this.state.currentIndex = -1;
        }
    }

    debounce(fn, delay = 300) {
        let timeoutId;
        return (...args) => {
            clearTimeout(timeoutId);
            timeoutId = window.setTimeout(() => fn(...args), delay);
        };
    }

    async fetchJson(path, { signal } = {}) {
        const url = this.resolveUrl(path);
        const response = await fetch(url, { signal });
        if (!response.ok) {
            const text = await response.text().catch(() => "");
            throw new Error(`Erreur ${response.status}: ${text || response.statusText}`);
        }
        return response.json();
    }

    setPlaybackStatus(message) {
        if (this.dom.status) {
            this.dom.status.textContent = message;
        }
    }

    setStats(message) {
        this.dom.stats.textContent = message;
    }

    setSearchBusy(busy) {
        this.dom.searchResults.setAttribute("aria-busy", busy ? "true" : "false");
    }

    renderPlaylist() {
        this.dom.playlistList.innerHTML = "";
        this.state.playlist.forEach((track, index) => {
            const clone = this.dom.playlistTemplate.content.firstElementChild.cloneNode(true);
            clone.dataset.index = String(index);
            clone.querySelector(".playlist-item__title").textContent = track.title ?? "Sans titre";
            clone.querySelector(".playlist-item__subtitle").textContent = [track.artist, track.album]
                .filter(Boolean)
                .join(" • ");
            clone.querySelector(".playlist-item__duration").textContent = this.formatTime(Number(track.length));
            const removeBtn = clone.querySelector(".playlist-item__remove");
            removeBtn.addEventListener("click", (event) => {
                event.stopPropagation();
                this.removeFromPlaylist(index);
            });
            clone.addEventListener("click", () => this.playTrack(index));
            if (index === this.state.currentIndex) {
                clone.classList.add("active");
            }
            this.dom.playlistList.appendChild(clone);
        });
        const count = this.state.playlist.length;
        this.dom.playlistCount.textContent = count === 0 ? "Aucune piste" : `${count} piste${count > 1 ? "s" : ""}`;
    }

    renderNowPlaying() {
        const track = this.state.playlist[this.state.currentIndex];
        if (!track) {
            this.dom.nowPlayingTitle.textContent = "Aucune piste";
            this.dom.nowPlayingArtist.textContent = "";
            this.dom.nowPlayingAlbum.textContent = "";
            this.toggleCoverVisibility(false);
            this.dom.totalTime.textContent = "0:00";
            this.dom.progress.value = "0";
            this.dom.playBtn.textContent = "▶";
            this.dom.playBtn.setAttribute("aria-label", "Lecture");
            this.setPlaybackStatus("En pause");
            return;
        }

        this.dom.nowPlayingTitle.textContent = track.title ?? "Sans titre";
        this.dom.nowPlayingArtist.textContent = track.artist ?? "";
        this.dom.nowPlayingAlbum.textContent = track.album ?? "";
        this.toggleCoverVisibility(Boolean(track.coverUrl));
        this.dom.totalTime.textContent = this.formatTime(Number(track.length));
        const isPaused = this.audio.isPaused();
        this.dom.playBtn.textContent = isPaused ? "▶" : "❚❚";
        this.dom.playBtn.setAttribute("aria-label", isPaused ? "Lecture" : "Pause");
        this.setPlaybackStatus(isPaused ? "En pause" : "Lecture");
    }

    renderSearchResults(results) {
        this.dom.searchResults.innerHTML = "";
        if (!results.length) {
            const empty = this.document.createElement("div");
            empty.className = "empty-state";
            empty.innerHTML = "<p>Aucune correspondance.</p><small>Essayez un autre mot-clé.</small>";
            this.dom.searchResults.appendChild(empty);
            return;
        }

        results.forEach((music) => {
            const clone = this.dom.searchTemplate.content.firstElementChild.cloneNode(true);
            clone.dataset.id = music.id;
            clone.querySelector(".result-item__title").textContent = music.title ?? "Sans titre";
            clone.querySelector(".result-item__subtitle").textContent = [music.artist, music.album]
                .filter(Boolean)
                .join(" • ");
            clone.querySelector(".add-btn").addEventListener("click", () => this.addToPlaylist(music));
            this.dom.searchResults.appendChild(clone);
        });
    }

    showSearchMessage(message, hint) {
        this.dom.searchResults.innerHTML = "";
        const empty = this.document.createElement("div");
        empty.className = "empty-state";
        empty.innerHTML = `<p>${message}</p>${hint ? `<small>${hint}</small>` : ""}`;
        this.dom.searchResults.appendChild(empty);
    }

    addToPlaylist(music) {
        const track = {
            id: music.id,
            title: music.title,
            artist: music.artist,
            album: music.album,
            length: music.length ?? music.time ?? music.duration,
            src: music.src,
            coverUrl: music.cover
        };
        this.state.playlist.push(track);
        if (this.state.currentIndex === -1) {
            this.state.currentIndex = 0;
        }
        this.saveState();
        this.renderPlaylist();
        if (this.state.playlist.length === 1) {
            this.playTrack(0);
        }
    }

    removeFromPlaylist(index) {
        if (index < 0 || index >= this.state.playlist.length) {
            return;
        }
        this.state.playlist.splice(index, 1);
        if (this.state.currentIndex === index) {
            if (this.state.playlist.length === 0) {
                this.state.currentIndex = -1;
                this.audio.pause();
                this.audio.clearSource();
            } else {
                this.state.currentIndex = Math.min(index, this.state.playlist.length - 1);
                this.playTrack(this.state.currentIndex, { autoplay: false });
            }
        } else if (this.state.currentIndex > index) {
            this.state.currentIndex -= 1;
        }
        this.saveState();
        this.renderPlaylist();
        this.renderNowPlaying();
    }

    clearPlaylist() {
        this.state.playlist = [];
        this.state.currentIndex = -1;
        this.audio.pause();
        this.audio.clearSource();
        this.saveState();
        this.renderPlaylist();
        this.renderNowPlaying();
    }

    shufflePlaylist() {
        for (let i = this.state.playlist.length - 1; i > 0; i -= 1) {
            const j = Math.floor(Math.random() * (i + 1));
            [this.state.playlist[i], this.state.playlist[j]] = [this.state.playlist[j], this.state.playlist[i]];
        }
        this.state.currentIndex = this.state.playlist.length > 0 ? 0 : -1;
        this.saveState();
        this.renderPlaylist();
        if (this.state.currentIndex !== -1) {
            this.playTrack(0, { autoplay: false });
        }
    }

    loadTrack(track) {
        this.audio.setSource(this.resolveUrl(track.src));
        this.toggleCoverVisibility(Boolean(track.coverUrl));
        this.audio.load();
    }
    toggleCoverVisibility(show) {
        if (!this.dom.cover) {
            return;
        }
        if (show) {
            this.dom.cover.classList.remove("is-hidden");
        } else {
            this.dom.cover.classList.add("is-hidden");
        }
    }

    loadTheme() {
        const savedTheme = localStorage.getItem(this.storageKeys.theme);
        if (savedTheme === "green") {
            this.applyTheme("green");
        } else {
            this.applyTheme("default");
        }
    }

    applyTheme(theme) {
        const body = this.document.body;
        const toggle = this.dom.themeToggle;
        const icon = toggle?.querySelector("span");
        if (theme === "green") {
            body.setAttribute("data-theme", "green");
            toggle?.setAttribute("aria-label", "Revenir au thème bleu");
            toggle?.setAttribute("title", "Revenir au thème bleu");
            if (icon) {
                icon.textContent = "🌿";
            }
        } else {
            body.removeAttribute("data-theme");
            toggle?.setAttribute("aria-label", "Activer le thème vert");
            toggle?.setAttribute("title", "Activer le thème vert");
            if (icon) {
                icon.textContent = "🎨";
            }
            theme = "default";
        }
        this.theme = theme;
        localStorage.setItem(this.storageKeys.theme, theme);
    }

    toggleTheme() {
        const nextTheme = this.theme === "green" ? "default" : "green";
        this.applyTheme(nextTheme);
    }

    playTrack(index, options = {}) {
        if (index < 0 || index >= this.state.playlist.length) {
            return;
        }
        const { autoplay = true } = options;
        this.state.currentIndex = index;
        const track = this.state.playlist[index];
        this.loadTrack(track);
        this.saveState();
        this.renderPlaylist();
        this.renderNowPlaying();
        if (autoplay) {
            this.audio
                .play()
                .then(() => {
                    this.dom.playBtn.textContent = "❚❚";
                    this.dom.playBtn.setAttribute("aria-label", "Pause");
                    this.setPlaybackStatus("Lecture");
                })
                .catch((err) => {
                    console.warn("Lecture impossible", err);
                    this.setPlaybackStatus("En pause (lecture bloquée)");
                });
        }
    }

    togglePlayPause() {
        if (!this.state.playlist.length) {
            return;
        }
        if (this.audio.isPaused()) {
            this.audio
                .play()
                .then(() => {
                    this.dom.playBtn.textContent = "❚❚";
                    this.dom.playBtn.setAttribute("aria-label", "Pause");
                    this.setPlaybackStatus("Lecture");
                })
                .catch((err) => {
                    console.warn("Lecture impossible", err);
                    this.setPlaybackStatus("En pause (lecture bloquée)");
                });
        } else {
            this.audio.pause();
            this.dom.playBtn.textContent = "▶";
            this.dom.playBtn.setAttribute("aria-label", "Lecture");
            this.setPlaybackStatus("En pause");
        }
    }

    playNext() {
        if (!this.state.playlist.length) {
            return;
        }
        const nextIndex = (this.state.currentIndex + 1) % this.state.playlist.length;
        this.playTrack(nextIndex);
    }

    playPrevious() {
        if (!this.state.playlist.length) {
            return;
        }
        const index = this.state.currentIndex - 1;
        this.playTrack(index >= 0 ? index : this.state.playlist.length - 1);
    }

    updateProgressBar() {
        const duration = this.audio.getDuration();
        if (duration) {
            const currentTime = this.audio.getCurrentTime();
            const percent = (currentTime / duration) * 100;
            this.dom.progress.value = percent.toString();
            this.dom.currentTime.textContent = this.formatTime(currentTime);
            this.dom.totalTime.textContent = this.formatTime(duration);
            return;
        }
        this.dom.progress.value = "0";
        this.dom.currentTime.textContent = "0:00";
        this.dom.totalTime.textContent = "0:00";
    }

    seek(event) {
        const duration = this.audio.getDuration();
        if (!duration) {
            return;
        }
        const percent = Number(event.target.value) / 100;
        this.audio.setCurrentTime(duration * percent);
    }

    updateVolume(event) {
        const value = Number(event.target.value);
        this.audio.setVolume(value);
        localStorage.setItem(this.storageKeys.volume, String(value));
    }

    async performSearch(query) {
        const trimmed = query.trim();
        if (trimmed.length < 2) {
            this.showSearchMessage("Pas de résultats");
            return;
        }

        if (this.state.searchAbortController) {
            this.state.searchAbortController.abort();
        }
        const controller = new AbortController();
        this.state.searchAbortController = controller;
        this.setSearchBusy(true);

        try {
            const results = await this.fetchJson(`search?term=${encodeURIComponent(trimmed)}&size=30`, {
                signal: controller.signal
            });
            this.renderSearchResults(results ?? []);
        } catch (err) {
            if (err.name === "AbortError") {
                return;
            }
            console.error("Erreur lors de la recherche", err);
            this.showSearchMessage("Impossible de charger les résultats.", "Vérifiez la connexion au serveur.");
        } finally {
            this.setSearchBusy(false);
        }
    }

    async updateLibraryStats() {
        try {
            const total = await fetch(this.resolveUrl("nbMusics")).then((res) => {
                if (!res.ok) {
                    throw new Error(res.statusText);
                }
                return res.text();
            });
            const number = Number(total);
            this.setStats(Number.isNaN(number) ? total : `${number.toLocaleString("fr-FR")} piste${number > 1 ? "s" : ""}`);
        } catch (err) {
            console.warn("Impossible de récupérer le nombre de musiques", err);
            this.setStats("Statistiques indisponibles");
        }
    }

    bindEvents() {
        this.dom.searchForm.addEventListener("submit", (event) => {
            event.preventDefault();
            this.performSearch(this.dom.searchInput.value);
        });

        this.dom.searchInput.addEventListener("input", (event) => {
            this.debouncedSearch(event.target.value);
        });

        this.dom.playBtn.addEventListener("click", () => this.togglePlayPause());
        this.dom.prevBtn.addEventListener("click", () => this.playPrevious());
        this.dom.nextBtn.addEventListener("click", () => this.playNext());
        this.dom.progress.addEventListener("input", (event) => this.seek(event));
        this.dom.volume.addEventListener("input", (event) => this.updateVolume(event));
        this.dom.clearPlaylistBtn.addEventListener("click", () => this.clearPlaylist());
        this.dom.shuffleBtn.addEventListener("click", () => this.shufflePlaylist());

        if (this.dom.themeToggle) {
            this.dom.themeToggle.addEventListener("click", () => this.toggleTheme());
        }

        this.audio.on("timeupdate", () => this.updateProgressBar());
        this.audio.on("loadedmetadata", () => {
            this.dom.totalTime.textContent = this.formatTime(this.audio.getDuration());
            this.updateProgressBar();
        });
        this.audio.on("ended", () => this.playNext());
        this.audio.on("play", () => {
            this.dom.playBtn.textContent = "❚❚";
            this.dom.playBtn.setAttribute("aria-label", "Pause");
            this.setPlaybackStatus("Lecture");
        });
        this.audio.on("pause", () => {
            this.dom.playBtn.textContent = "▶";
            this.dom.playBtn.setAttribute("aria-label", "Lecture");
            this.setPlaybackStatus("En pause");
        });
        this.audio.on("error", () => {
            this.setPlaybackStatus("Erreur de lecture");
        });

        window.addEventListener("keydown", (event) => {
            if (event.code === "Space" && (event.target === this.document.body || event.target === this.dom.searchInput)) {
                event.preventDefault();
                this.togglePlayPause();
            }
            if (event.code === "ArrowRight" && !this.audio.isPaused()) {
                const duration = this.audio.getDuration();
                const current = this.audio.getCurrentTime();
                const target = duration ? Math.min(current + 5, duration) : current + 5;
                this.audio.setCurrentTime(target);
            }
            if (event.code === "ArrowLeft" && !this.audio.isPaused()) {
                const current = this.audio.getCurrentTime();
                this.audio.setCurrentTime(Math.max(current - 5, 0));
            }
        });
    }
}

document.addEventListener("DOMContentLoaded", () => {
    const app = new MusicSpaApp();
    app.init();
    window.musicSpaApp = app;
});
