function formatTime(value) {
    if (Number.isNaN(value) || value == null) {
        return "0:00";
    }
    const seconds = Math.floor(value);
    const minutes = Math.floor(seconds / 60);
    const rest = seconds % 60;
    return `${minutes}:${rest < 10 ? "0" : ""}${rest}`;
}

// Interface which manage playlist
class IPlaylist {
    add(music) {
        return Promise.resolve();
    }

    remove(index) {
    }

    clear() {
    }
}

const getPlaylistController = (isRemote = false, arg = {}) => {
    return isRemote ? new RemotePlaylistController(arg) : new PlaylistController();
}

class RemotePlaylistController extends IPlaylist {
    constructor(remoteController) {
        super();
        this.remoteController = remoteController;
    }

    add(music) {
        // Add to remote and wait result
        return this.remoteController.addMusic(music.id)
    }

    remove(index) {
        this.remoteController.removeMusic(index);
    }

    clear() {
        this.remoteController.clearPlaylist();
    }
}

class PlaylistController extends IPlaylist {
    constructor() {
        super()
    }
}

class IController {
    showStartPlayer() {
    }

    setSource(url, index) {
    }

    clearSource() {
    }

    setVolume(value) {
    }

    isPaused() {
    }

    pause() {
    }

    getDuration() {
        return 0;
    }

    play() {
    }

    unpause() {
    }

    updateProgress(fctDisplay) {
    }
}

class RemoteAudioController extends IController {
    constructor(remote, fctDisplayer, music) {
        super();
        this.remote = remote;
        this._isPaused = true;
        this.fctDisplayer = fctDisplayer;
        this.music = music;
    }

    setSource(url, index) {
        this.currentIndex = index;
    }

    isPaused() {
        return this._isPaused;
    }

    pause() {
        if (this._isPaused) {
            return;
        }
        if (this.updateTime != null) {
            clearInterval(this.updateTime)
            this.updateTime = null;
        }
        this.updateProgress();
        this._isPaused = true;
        return this.remote.pauseMusic()
    }

    showStartPlayer(position, length, isPaused) {
        this._isPaused = isPaused;
        this.position = position;
        this.length = length;
        this._showProgress();
        if (!isPaused) {
            this._startPlayer();
        }
    }

    unpause() {
        this._isPaused = false;
        this._startPlayer();
        return this.remote.unpauseMusic();
    }

    play() {
        this._isPaused = false;
        return this.remote.playMusic(this.currentIndex)
            .then(()=>setTimeout(()=>this.updateProgress().then(() => this._startPlayer()),500));
    }

    _startPlayer() {
        if (this.updateTime) {
            clearInterval(this.updateTime);
        }
        this.music.showPlayingStatus(false)
        this.updateTime = setInterval(() => {
            this.position += 1000;
            if(!this._showProgress()){
                clearInterval((this.updateTime))
                this.updateTime = null;
            }
        }, 1000)
    }

    setVolume(value) {
        value === 1 ? this.remote.increaseVolume() : this.remote.decreaseVolume()
    }

    updateProgress() {
        return this.remote.getState().then(data => {
            this.position = data.position;
            this.length = data.length;
            this._showProgress()
        });
    }

    _showProgress() {
        if(this.position > this.length){
            return false;
        }
        this.fctDisplayer((this.position / this.length) * 100, formatTime(this.position / 1000), formatTime(this.length / 1000))
        return true;
    }
}

class AudioController extends IController {
    constructor(audioElement) {
        super();
        this.element = audioElement;
    }

    updateProgress(fctDisplay) {
        const duration = this.getDuration();
        if (!duration) {
            return fctDisplay("0", "0:00", "0:00")
        }
        const currentTime = this.getCurrentTime();
        const percent = (currentTime / duration) * 100;
        fctDisplay(percent.toString(), formatTime(currentTime), formatTime(duration))
    }

    setSource(url, index) {
        this.element.src = url;
        this.element.load();
    }

    clearSource() {
        this.element.removeAttribute("src");
    }

    play() {
        return this.element.play();
    }

    pause() {
        return this.element.pause();
    }

    unpause() {
        return this.element.play();
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

class RemotePlayerController extends IPlaylist {
    constructor({Name, Id}, musicApp) {
        super();
        this.music = musicApp
        this.sse = new EventSource(requester.resolveUrl(`/share?id=${Id}&device=mobile`));
        this.sse.addEventListener('id', data => {
            this.playerId = Id;
            this.sessionId = data.data;
        })
        this.sse.addEventListener('playlist', message => {
            this.loadPlaylist(JSON.parse(message.data));
        })
        this.sse.addEventListener('error', ii => {
            console.log("ERROR", ii)
        })

        this.sse.addEventListener('message', ii => {
            console.log("MESSAGE", ii)
        })
        this.sse.addEventListener('close', a => {
            console.log("GO", a)
        })
    }

    playMusic(index) {
        return requester.simpleFetch(`shareUpdate?event=playMusic&id=${this.playerId}&data={"position":${index}}`)
    }

    pauseMusic() {
        return requester.simpleFetch(`shareUpdate?event=pause&id=${this.playerId}`)
    }

    unpauseMusic() {
        return requester.simpleFetch(`shareUpdate?event=play&id=${this.playerId}`)
    }

    increaseVolume() {
        return requester.simpleFetchAsBool(`shareUpdate?event=volumeUp&id=${this.playerId}`)
    }

    decreaseVolume() {
        return requester.simpleFetchAsBool(`shareUpdate?event=volumeDown&id=${this.playerId}`)
    }

    addMusic(musicId) {
        return requester.simpleFetch(`shareUpdate?event=add&id=${this.playerId}&data=${musicId}`, {method: 'POST'})
    }

    removeMusic(index) {
        return requester.simpleFetchAsBool(`shareUpdate?event=remove&id=${this.playerId}&index=${index}`)
    }

    clearPlaylist() {
        return requester.simpleFetchAsBool(`shareUpdate?event=cleanPlaylist&id=${this.playerId}`)
    }

    getState() {
        return requester.fetch(`sendShareRequest?event=state&id=${this.playerId}`);
    }

    loadPlaylist(details) {
        this.music.clearPlaylist(false)
        this.music.audio.showStartPlayer(details.position, details.length, !details.playing)
        requester.fetch(`musicsInfo?ids=[${details.ids}]`).then(data => {
            data.forEach(m => this.music.state.add(m))
            this.music.renderPlaylist()
            this.music.setCurrent(details.current, details.playing)
        })
    }
}

class Requester {
    constructor() {
        this.apiBase = document.location.href.replace("mobile/", "");
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

    async fetch(path, isJson = true, {signal} = {}) {
        const url = this.resolveUrl(path);
        const response = await fetch(url, {signal});
        if (!response.ok) {
            const text = await response.text().catch(() => "");
            throw new Error(`Erreur ${response.status}: ${text || response.statusText}`);
        }
        return isJson ? response.json() : response.text()
    }

    simpleFetchAsBool(path, args = {}) {
        try {
            this.simpleFetch(path, args)
        } catch (e) {
            return false;
        }
        return true;
    }

    async simpleFetch(path, args = {}) {
        const url = this.resolveUrl(path);
        const response = await fetch(url, args);
        if (!response.ok) {
            const text = await response.text().catch(() => "");
            throw new Error(`Erreur ${response.status}: ${text || response.statusText}`);
        }
        return response;
    }
}

class StateManager {
    constructor() {
        this.playlist = [];
        this.currentIndex = -1;
        this.volume = 0;
        this.searchAbortController = null;
        this.enableStorage = true;
        this.storageKeys = {
            playlist: "music-server-mobile:playlist",
            currentIndex: "music-server-mobile:current-index",
            volume: "music-server-mobile:volume",
            theme: "music-server-mobile:theme"
        };
    }

    setStorage(isEnable) {
        this.enableStorage = isEnable;
    }

    isEmpty() {
        return this.playlist.length === 0;
    }

    size() {
        return this.playlist.length;
    }

    getCurrent() {
        return this.currentIndex;
    }

    remove(index) {
        this.playlist.splice(index, 1);
    }

    clear() {
        this.playlist = [];
        this.currentIndex = -1;
    }

    clearPlaylist() {
        this.playlist = [];
    }

    get(index = this.currentIndex) {
        if (index >= 0 && index < this.playlist.length) {
            return this.playlist[index]
        }
        return null
    }

    add(track) {
        this.playlist.push(track)
        if (this.currentIndex === -1) {
            this.currentIndex = 0;
        }
    }

    next() {
        return (this.currentIndex + 1 % this.size())
    }

    previous() {
        return this.currentIndex - 1 < 0 ? this.size() - 1 : this.currentIndex - 1;
    }

    save() {
        if (this.enableStorage) {
            localStorage.setItem(this.storageKeys.playlist, JSON.stringify(this.playlist));
            localStorage.setItem(this.storageKeys.currentIndex, String(this.currentIndex));
            localStorage.setItem(this.storageKeys.theme, this.theme ?? '')
        }
    }

    load() {
        try {
            const rawPlaylist = localStorage.getItem(this.storageKeys.playlist) || "[]";
            const playlist = JSON.parse(rawPlaylist);
            if (Array.isArray(playlist)) {
                this.playlist = playlist;
            }
            const index = parseInt(localStorage.getItem(this.storageKeys.currentIndex) ?? "-1", 10);
            if (!Number.isNaN(index)) {
                this.currentIndex = index;
            }
            const volume = parseFloat(localStorage.getItem(this.storageKeys.volume) ?? "0.7");
            this.volume = Number.isNaN(volume) ? 0.7 : Math.min(Math.max(volume, 0), 1);
            this.theme = localStorage.getItem(this.storageKeys.theme)
        } catch (err) {
            console.warn("Impossible de charger l'état du lecteur", err);
            this.playlist = [];
            this.currentIndex = -1;
        }
    }
}

const requester = new Requester();
const kindFolder = {artist:1,album:2};

class MusicSpaApp {
    constructor(doc = document) {
        this.document = doc;
        this.root = doc.documentElement;
        this.playlistController = getPlaylistController();

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
            volumeUp: doc.getElementById("vup-btn"),
            volumeDown: doc.getElementById("vdown-btn"),
            searchForm: doc.getElementById("search-form"),
            searchInput: doc.getElementById("search-input"),
            searchResults: doc.getElementById("search-results"),
            playlistList: doc.getElementById("playlist-list"),
            playlistCount: doc.getElementById("playlist-count"),
            clearPlaylistBtn: doc.getElementById("clear-playlist-btn"),
            shuffleBtn: doc.getElementById("shuffle-btn"),
            audio: doc.getElementById("audio-player"),
            searchTemplate: doc.getElementById("search-result-template"),
            searchFolderTemplate: doc.getElementById("search-folder-result-template"),
            playlistTemplate: doc.getElementById("playlist-item-template"),
            themeToggle: doc.getElementById("theme-toggle")
        };

        this.audio = new AudioController(this.dom.audio);
        this.playlist = new PlaylistController();
        this.state = new StateManager();
        this.debouncedSearch = this.debounce((value) => this.performSearch(value), 350);
    }

    init() {
        this.loadState();
        this.loadTheme();
        this.initShare();
        this.renderPlaylist();
        this.renderNowPlaying();
        this.bindEvents();
        void this.updateLibraryStats();
        this.loadCurrentTrack();
    }

    async initShare() {
        if (await this.isShareExists()) {
            const share = document.getElementById('connection-share');
            share.style.setProperty("display", "")
            share.addEventListener('click', () => this.connectToShare())
        }
    }

    isShareExists() {
        return requester.fetch("shares").then(data => data.length === 1)
    }

    connectToShare() {
        // Get all shares, if only one, connect to it, otherwise, error message
        requester.fetch("shares").then(data => {
            this.state.setStorage(false);
            this.remotePlayer = new RemotePlayerController(data[0], this);
            this.playlistController = getPlaylistController(true, this.remotePlayer);
            this.dom.volume.style.setProperty("display", "none");
            this.dom.volumeUp.style.setProperty("display", "");
            this.dom.volumeDown.style.setProperty("display", "");
            this.audio = new RemoteAudioController(this.remotePlayer, (a, b, c) => this._updateProgressBarDisplay(a, b, c),this);
        })
    }

    loadCurrentTrack() {
        const track = this.state.get();
        if (track && track.src) {
            this.loadTrack(track);
        }
    }

    loadState() {
        this.state.load();
        this.audio.setVolume(this.state.volume);
        this.dom.volume.value = String(this.audio.getVolume());
    }

    debounce(fn, delay = 300) {
        let timeoutId;
        return (...args) => {
            clearTimeout(timeoutId);
            timeoutId = window.setTimeout(() => fn(...args), delay);
        };
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
            clone.querySelector(".playlist-item__duration").textContent = formatTime(Number(track.length));
            const removeBtn = clone.querySelector(".playlist-item__remove");
            removeBtn.addEventListener("click", (event) => {
                event.stopPropagation();
                this.removeFromPlaylist(index);
            });
            clone.addEventListener("click", () => this.playTrack(index));
            if (index === this.state.getCurrent()) {
                clone.classList.add("active");
            }
            this.dom.playlistList.appendChild(clone);
        });
        const count = this.state.size();
        this.dom.playlistCount.textContent = count === 0 ? "Aucune piste" : `${count} piste${count > 1 ? "s" : ""}`;
    }

    renderNowPlaying() {
        const track = this.state.get();
        if (!track) {
            this.dom.nowPlayingTitle.textContent = "Aucune piste";
            this.dom.nowPlayingArtist.textContent = "";
            this.dom.nowPlayingAlbum.textContent = "";
            this.toggleCoverVisibility(false);
            this.dom.currentTime.textContent = "0:00";
            this.dom.totalTime.textContent = "0:00";
            this.dom.progress.value = "0";
            this.showPlayingStatus(true)
            return;
        }

        this.dom.nowPlayingTitle.textContent = track.title ?? "Sans titre";
        this.dom.nowPlayingArtist.textContent = track.artist ?? "";
        this.dom.nowPlayingAlbum.textContent = track.album ?? "";
        this.toggleCoverVisibility(Boolean(track.coverUrl));
        //this.dom.totalTime.textContent = formatTime(Number(track.length));
        this.showPlayingStatus(this.audio.isPaused())
    }

    showPlayingStatus(isPaused) {
        if (isPaused) {
            this.dom.playBtn.textContent = "▶";
            this.dom.playBtn.setAttribute("aria-label", "Lecture");
            this.setPlaybackStatus("En pause");
        } else {
            this.dom.playBtn.textContent = "❚❚";
            this.dom.playBtn.setAttribute("aria-label", "Pause");
            this.setPlaybackStatus("Lecture");
        }
    }

    renderSearchResults(results, isFolderView = false, kind = kindFolder.album) {
        this.dom.searchResults.innerHTML = "";
        if (!results.length) {
            const empty = this.document.createElement("div");
            empty.className = "empty-state";
            empty.innerHTML = "<p>Aucune correspondance.</p><small>Essayez un autre mot-clé.</small>";
            this.dom.searchResults.appendChild(empty);
            return;
        }
        results.forEach((music) => {
            this.dom.searchResults.appendChild(isFolderView ? this.showFolder(music, kind) : this.showMusic(music));
        });
    }
    adaptMusic(music){
        music.title = music.title ?? music.name ?? "Sans titre";
        music.album = music.album ?? music.infos.album;
        music.artist = music.artist ?? music.infos.artist;
        return music;
    }
    showMusic(music){
        music = this.adaptMusic(music);
        const clone = this.dom.searchTemplate.content.firstElementChild.cloneNode(true);
        clone.dataset.id = music.id;
        clone.querySelector(".result-item__title").textContent = music.title;
        clone.querySelector(".result-item__subtitle").textContent = [music.artist, music.album]
            .filter(Boolean)
            .join(" • ");
        clone.querySelector(".add-btn").addEventListener("click", () => this.addToPlaylist(music));
        return clone;
    }

    showFolder(folder, kind){
        const clone = this.dom.searchFolderTemplate.content.firstElementChild.cloneNode(true);
        clone.dataset.id = folder.id;
        clone.querySelector(".result-item__title").textContent = folder.name ?? "Sans titre";
        if(kind === kindFolder.album) {
            clone.querySelector(".show-btn").addEventListener("click", () => this.showMusicOfAlbum(folder.id));
            clone.querySelector(".add-btn").addEventListener("click", () => this.addMusicsOfAlbum(folder.id));
        }else{
            clone.querySelector(".show-btn").addEventListener("click", () => this.showMusicOfArtist(folder.id));
            clone.querySelector(".add-btn").addEventListener("click", () => this.addMusicsOfArtist(folder.id));
        }
        return clone;

    }

    addMusicsOfArtist(id){
        this.addMusicsFromUrl(`listByArtist?id=${id}&detail=true`)
    }

    addMusicsOfAlbum(id){
        this.addMusicsFromUrl(`listByOnlyAlbums?idAlbum=${id}`)
    }

    addMusicsFromUrl(url){
        requester.fetch(`${url}&size=60`, true).then(results=>{
            results.forEach(m => {
                this.addToPlaylist(this.adaptMusic(m))
            })
            this.state.save();
            this.renderPlaylist()
        });
    }

    showMusicOfAlbum(id){
        requester.fetch(`listByOnlyAlbums?idAlbum=${id}&size=60`, true).then(results=>this.renderSearchResults(results ?? []));
    }
    showMusicOfArtist(id){
        requester.fetch(`listByArtist?id=${id}&size=60`, true).then(results=>this.renderSearchResults(results ?? []));
    }

    showSearchMessage(message, hint) {
        this.dom.searchResults.innerHTML = "";
        const empty = this.document.createElement("div");
        empty.className = "empty-state";
        empty.innerHTML = `<p>${message}</p>${hint ? `<small>${hint}</small>` : ""}`;
        this.dom.searchResults.appendChild(empty);
    }

    trackFromMusic(music) {
        return {
            id: music.id,
            title: music.title,
            artist: music.artist,
            album: music.album,
            length: parseInt(music.length) ?? parseInt(music.time) ?? parseInt(music.duration) ?? music.infos != null ? parseInt(music.infos.time):0,
            src: music.src || music.infos != null ? music.infos.src : '',
            coverUrl: music.cover
        };
    }

    addToPlaylist(music) {
        const track = this.trackFromMusic(music);
        this.playlistController.add(track).then(() => {
            this.state.add(track);
            this.state.save()
            this.renderPlaylist();
            if (this.state.size() === 1) {
                this.playTrack(0);
            }
        });
    }

    removeFromPlaylist(index) {
        if (index < 0 || index >= this.state.size()) {
            return;
        }
        this.playlistController.remove(index)
        this.state.remove(index);
        if (this.state.getCurrent() === index) {
            if (this.state.isEmpty()) {
                this.state.currentIndex = -1;
                this.audio.pause();
                this.audio.clearSource();
            } else {
                this.state.currentIndex = this.state.previous();
                this.playTrack(this.state.getCurrent(), {autoplay: false});
            }
        } else if (this.state.getCurrent() > index) {
            this.state.currentIndex = this.state.previous();
        }
        this.state.save();
        this.renderPlaylist();
        this.renderNowPlaying();
    }

    clearPlaylist(propagation = true) {
        this.state.clear();
        if (propagation) {
            this.playlistController.clear();
        }
        this.audio.pause();
        this.audio.clearSource();
        this.state.save();
        this.renderPlaylist();
        this.renderNowPlaying();
    }

    shufflePlaylist() {
        for (let i = this.state.size() - 1; i > 0; i -= 1) {
            const j = Math.floor(Math.random() * (i + 1));
            [this.state.playlist[i], this.state.playlist[j]] = [this.state.playlist[j], this.state.playlist[i]];
        }
        this.state.currentIndex = this.state.isEmpty() ? -1 : 0;
        this.state.save();
        this.renderPlaylist();
        if (this.state.getCurrent() !== -1) {
            this.playTrack(0, {autoplay: false});
        }
    }

    loadTrack(track, index) {
        this.audio.setSource(requester.resolveUrl(track.src), index);
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
        this.applyTheme(this.state.theme === "green" ? "green" : "default")
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
        this.state.theme = theme;
        this.state.save()
    }

    toggleTheme() {
        const nextTheme = this.state.theme === "green" ? "default" : "green";
        this.applyTheme(nextTheme);
    }

    setCurrent(index) {
        if (index === -1) {
            return;
        }
        this.state.currentIndex = index;
        this.renderNowPlaying();
    }

    playTrack(index, options = {}) {
        if (this.state.size() < 0 || index < 0 || index >= this.state.size()) {
            return;
        }
        const {autoplay = true} = options;

        const track = this.state.get(index);
        this.loadTrack(track, index);
        this.setCurrent(index);
        this.renderPlaylist();

        this.state.save();
        if (autoplay) {
            this.audio
                .play()
                .then(() => {
                    this.audio.updateProgress((a, b, c) => this._updateProgressBarDisplay(a, b, c))
                    this.showPlayingStatus(false)
                })
                .catch((err) => {
                    console.warn("Lecture impossible", err);
                    this.setPlaybackStatus("En pause (lecture bloquée)");
                });
        }
    }

    togglePlayPause() {
        if (!this.state.size()) {
            return;
        }
        if (this.audio.isPaused()) {
            this.audio
                .unpause()
                .then(() => {
                    this.showPlayingStatus(false)
                })
                .catch((err) => {
                    console.warn("Lecture impossible", err);
                    this.setPlaybackStatus("En pause (lecture bloquée)");
                });
        } else {
            this.audio.pause();
            this.showPlayingStatus(true)
        }
    }

    playNext() {
        this.playTrack(this.state.next());
    }

    playPrevious() {
        this.playTrack(this.state.previous());
    }

    _updateProgressBarDisplay(progress, current, total) {
        this.dom.progress.value = progress;
        this.dom.currentTime.textContent = current;
        this.dom.totalTime.textContent = total;
    }

    updateProgressBar() {
        this.audio.updateProgress((a, b, c) => this._updateProgressBarDisplay(a, b, c))
    }

    seek(event) {
        const duration = this.audio.getDuration();
        if (!duration) {
            return;
        }
        const percent = Number(event.target.value) / 100;
        this.audio.setCurrentTime(duration * percent);
    }

    increaseVolume() {
        this.audio.setVolume(1)
    }

    decreaseVolume() {
        this.audio.setVolume(-1)
    }

    updateVolume(event) {
        const value = Number(event.target.value);
        this.audio.setVolume(value);
        this.state.volume = value;
        this.state.save();
    }

    async performSearch(query) {
        const trimmed = query;//.trim();
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
        if(trimmed.indexOf(":") === 0){
            if(trimmed.indexOf(":album ") === 0 || trimmed.indexOf(":al ") === 0){
                const results = await requester.fetch(`albums?term=${encodeURIComponent(trimmed.replace(/(:album )|(:al )/,""))}&size=30`, true, {
                    signal: controller.signal
                });
                return this.renderSearchResults(results ?? [], true, kindFolder.album);
            }
            if(trimmed.indexOf(":artist ") === 0 || trimmed.indexOf(":ar ") === 0){
                const results = await requester.fetch(`artists?artist=${encodeURIComponent(trimmed.replace(/(:artist )|(:ar )/,""))}&size=30`, true, {
                    signal: controller.signal
                });
                return this.renderSearchResults(results ?? [], true, kindFolder.artist);
            }
            return
        }

        try {
            const results = await requester.fetch(`search?term=${encodeURIComponent(trimmed)}&size=30`, true, {
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
            const total = await fetch(requester.resolveUrl("nbMusics")).then((res) => {
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
        this.dom.volumeUp.addEventListener("click", (_) => this.increaseVolume());
        this.dom.volumeDown.addEventListener("click", (_) => this.decreaseVolume());
        this.dom.clearPlaylistBtn.addEventListener("click", () => this.clearPlaylist());
        this.dom.shuffleBtn.addEventListener("click", () => this.shufflePlaylist());

        if (this.dom.themeToggle) {
            this.dom.themeToggle.addEventListener("click", () => this.toggleTheme());
        }

        this.audio.on("timeupdate", () => this.updateProgressBar());
        this.audio.on("loadedmetadata", () => {
            this.dom.totalTime.textContent = formatTime(this.audio.getDuration());
            this.updateProgressBar();
        });
        this.audio.on("ended", () => this.playNext());
        this.audio.on("play", () => {
            this.showPlayingStatus(false);
        });
        this.audio.on("pause", () => {
            this.showPlayingStatus(true);
        });
        this.audio.on("error", () => {
            this.setPlaybackStatus("Erreur de lecture");
        });

        window.addEventListener("keydown", (event) => {
            if (event.code === "Space" && event.target !== this.dom.searchInput) {
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
