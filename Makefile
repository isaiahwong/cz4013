default:
	make run

run:
	make -C ./server 

gen_flights:
	python3 scripts/seed_gen.py && mv flights.csv ./server