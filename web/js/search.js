document.addEventListener('DOMContentLoaded', function() {
    document.getElementById('search-form').addEventListener('submit', function(event) {
        event.preventDefault(); 
    });

    function filterArtists() {
        const query = document.getElementById('search-query').value;
        const dates = document.getElementById('dates').value;
        const memberCount = document.getElementById('memberCount').value;
    
        fetch(`/artists?q=${encodeURIComponent(query)}&dates=${encodeURIComponent(dates)}&memberCount=${encodeURIComponent(memberCount)}`)
            .then(response => {
                if (!response.ok) {
                    throw new Error(`HTTP error! status: ${response.status}`);
                }
                return response.text();
            })
            .then(text => {
                console.log('Raw response:', text); 
                try {
                    return JSON.parse(text); 
                } catch (e) {
                    throw new Error(`Invalid JSON: ${e.message}`);
                }
            })
            .then(data => updateArtistList(data))
            .catch(error => console.error('Error parsing JSON:', error));
    }

    function updateArtistList(artists) {
        const list = document.querySelector('ul');
        list.innerHTML = '';

        artists.forEach(artist => {
            const listItem = document.createElement('li');
            listItem.innerHTML = `
                <h2>${artist.Name}</h2>
                <img src="${artist.Image}" alt="Image of ${artist.Name}" width="150">
                <!-- Autres détails -->
            `;
            list.appendChild(listItem);
        });
    }
});