package chef

import (
	"net/http"
	"net/url"

	"github.com/sauerbraten/chef/db"
	"github.com/sauerbraten/chef/ips"
)

func parseLookupRequest(resp http.ResponseWriter, req *http.Request) (
	nameOrIP string,
	sorting db.Sorting,
	last90DaysOnly, directLookupForced bool,
	redirected bool,
) {
	nameOrIP = req.FormValue("q")

	sorting = db.ByNameFrequency
	if req.FormValue("sorting") == db.ByLastSeen.Identifier {
		sorting = db.ByLastSeen
	}

	last90DaysOnly = !(req.FormValue("search_old") == "true")

	directLookupForced = req.FormValue("direct") == "true"

	// (permanently) redirect partial IP queries
	if ips.IsPartialOrFullCIDR(nameOrIP) {
		subnet := ips.GetSubnet(nameOrIP)

		if nameOrIP != subnet.String() {
			u, _ := url.ParseRequestURI(req.RequestURI) // it's safe to assume this will not fail
			params := u.Query()
			params.Set("q", subnet.String())
			u.RawQuery = params.Encode()
			http.Redirect(resp, req, u.String(), http.StatusPermanentRedirect)
			redirected = true
		}
	}

	return
}
