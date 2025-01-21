// Make sure the DOM is fully loaded before manipulating elements
document.addEventListener('DOMContentLoaded', () => {
    const searchInput = document.getElementById('search-query');
    const artistList = document.querySelector('ul');
    // Update the search of the user
    searchInput.addEventListener('input', () => {
        const query = searchInput.value;
        // Get artist data
        fetch(`/artists?q=${encodeURIComponent(query)}`)
            .then(response => response.text())
            .then(html => {
                const parser = new DOMParser();
                const doc = parser.parseFromString(html, 'text/html');
                const newArtistList = doc.querySelector('ul');
                artistList.innerHTML = newArtistList.innerHTML;
            })
            // If there is a problem, then print this error
            .catch(err => {
                console.error('Error fetching artists:', err);
            });
    });
});