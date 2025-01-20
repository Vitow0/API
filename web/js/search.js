document.addEventListener('DOMContentLoaded', () => {
    const searchInput = document.getElementById('search-query');
    const artistList = document.querySelector('ul');

    searchInput.addEventListener('input', () => {
        const query = searchInput.value;

        fetch(`/artists?q=${encodeURIComponent(query)}`)
            .then(response => response.text())
            .then(html => {
                const parser = new DOMParser();
                const doc = parser.parseFromString(html, 'text/html');
                const newArtistList = doc.querySelector('ul');
                artistList.innerHTML = newArtistList.innerHTML;
            })
            .catch(err => {
                console.error('Error fetching artists:', err);
            });
    });
});