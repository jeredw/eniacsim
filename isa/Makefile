all: sixty.html sixty.e
sixty.html: sixty.nw
	noweave -index -html -n sixty.nw > sixty.html
sixty.e: sixty.nw
	notangle sixty.nw > sixty.e
