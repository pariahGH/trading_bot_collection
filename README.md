## Trading Bot Collection

This project was intended to collect trading data from Binance, apply some TA formulas to it, and stick it all in a TimescaleDB instance as well as pushing it over Redis PubSub for consumption by strategy/trade management processes. 

Everything works except for indicator generation. For some reason, probably RAM related, the postgres processes would trigger the OOM killer - so I buffered updates and opened/closed a connection for every few minutes worth of data. 

Found out it was a lot easier to just trade news.

This originally became a thing as I needed a good project to learn Go with - this and the corresponding strategy/trade management engine were an excellent learning experience. 