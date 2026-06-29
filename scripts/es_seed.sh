#!/bin/bash
# Seed Elasticsearch with content from Postgres (run after docker compose up)
ES_URL="${ES_URL:-http://localhost:9200}"
INDEX="content"

echo "Seeding Elasticsearch index: $INDEX"

seed_doc() {
  curl -s -X PUT "$ES_URL/$INDEX/_doc/$1" \
    -H "Content-Type: application/json" \
    -d "$2" > /dev/null
  echo "indexed: $3"
}

seed_doc "a1b2c3d4-0001-0001-0001-000000000001" '{
  "id":"a1b2c3d4-0001-0001-0001-000000000001","title":"The Dark Knight",
  "description":"Batman faces the Joker in Gotham City.","type":"movie",
  "release_year":2008,"genres":["Action","Thriller"],"view_count":9800,"rating_avg":9.2,
  "thumbnail_url":"https://placeholder.co/300x450",
  "title_suggest":{"input":["The Dark Knight","Dark Knight","Batman"]}
}' "The Dark Knight"

seed_doc "a1b2c3d4-0002-0002-0002-000000000002" '{
  "id":"a1b2c3d4-0002-0002-0002-000000000002","title":"Inception",
  "description":"A thief enters dreams to plant an idea.","type":"movie",
  "release_year":2010,"genres":["Sci-Fi","Thriller"],"view_count":8700,"rating_avg":8.8,
  "thumbnail_url":"https://placeholder.co/300x450",
  "title_suggest":{"input":["Inception","Dream Heist"]}
}' "Inception"

seed_doc "a1b2c3d4-0003-0003-0003-000000000003" '{
  "id":"a1b2c3d4-0003-0003-0003-000000000003","title":"Breaking Bad",
  "description":"A chemistry teacher turns to making meth.","type":"series",
  "release_year":2008,"genres":["Drama","Thriller"],"view_count":15000,"rating_avg":9.5,
  "thumbnail_url":"https://placeholder.co/300x450",
  "title_suggest":{"input":["Breaking Bad","Walter White"]}
}' "Breaking Bad"

seed_doc "a1b2c3d4-0004-0004-0004-000000000004" '{
  "id":"a1b2c3d4-0004-0004-0004-000000000004","title":"Interstellar",
  "description":"Astronauts travel through a wormhole near Saturn.","type":"movie",
  "release_year":2014,"genres":["Sci-Fi","Drama"],"view_count":7200,"rating_avg":8.6,
  "thumbnail_url":"https://placeholder.co/300x450",
  "title_suggest":{"input":["Interstellar","Space Travel"]}
}' "Interstellar"

seed_doc "a1b2c3d4-0005-0005-0005-000000000005" '{
  "id":"a1b2c3d4-0005-0005-0005-000000000005","title":"Stranger Things",
  "description":"Kids in a small town face supernatural forces.","type":"series",
  "release_year":2016,"genres":["Horror","Sci-Fi"],"view_count":12000,"rating_avg":8.7,
  "thumbnail_url":"https://placeholder.co/300x450",
  "title_suggest":{"input":["Stranger Things","Upside Down"]}
}' "Stranger Things"

echo "Done!"
