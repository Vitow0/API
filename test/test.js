const fetchGeocode = async (address) => {
    const url = `https://google-map-places.p.rapidapi.com/maps/api/geocode/json?address=${encodeURIComponent(address)}`;
    const options = {
      method: 'GET',
      headers: {
        'x-rapidapi-key': '7a2cfcfda4msh2f03e4de2794082p1b4d77jsnac469a73d4b2',
        'x-rapidapi-host': 'google-map-places.p.rapidapi.com',
      },
    };
  
    try {
      const response = await fetch(url, options);
      if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
      const result = await response.json();
  
      if (result.status === 'OK') {
        result.results.forEach((res) => {
          console.log(`Adresse : ${res.formatted_address}`);
          console.log(`Latitude : ${res.geometry.location.lat}`);
          console.log(`Longitude : ${res.geometry.location.lng}`);
        });
      } else {
        console.error(`Erreur API : ${result.status}`);
      }
    } catch (error) {
      console.error('Erreur Fetch :', error);
    }
  };
  
  fetchGeocode('1600 Amphitheatre Parkway, Mountain View, CA');
  