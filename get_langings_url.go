package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

var urlForGetLandings string = "https://tools.kontur.ru/module/LandingPage/list"

type Payload struct {
	SelectItems []struct {
		Value string `json:"value"`
		Text  string `json:"text"`
	} `json:"selectItems"`
	Type string `json:"type"`
}

var cookie string

func GetLandingsUrls() ([]string, error) {

	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
	}
	client := &http.Client{
		Transport: tr,
	}

	req, err := http.NewRequest("GET", urlForGetLandings, nil)

	cookie = "ngtoken=LhHOD2ftZlsePshoCrhhAg==; sid=36453859; token=djnDp40Uq44g1Mr0fBd5tjYIz6G1HBQE6+tiijkxwn0xv/EFj+zVdwIHQyFu6+k0HvMEJEWXV4ME3Mvo84Nk6WbkPzz7FL/5TPIzo9azFSGASceHY7hq4HEFQM0W6aey; auth.sid=66C648A6E0A4AA4FBE991A670AE377446A874B9699C0154CB174871B9AC3819E; _ top100_id=t1.7709112.547167062.1752829980391; t3_sid_7709112=s1.1885937431.1753262271490.1753262271496.2.3.0.0; adrdel=1755067722324; adrcid=A_DeIilUVz3vWIstGSgAXfQ; acs_3=%7B%22hash%22%3A%221aa3f9523ee6c2690cb34fc702d4143056487c0d%22%2C%22nst%22%3A1755160854598%2C%22sl%22%3A%7B%22224%22%3A1755074454598%2C%221228%22%3A1755074454598%7D%7D; gnezdo_uid=uZQlT2ia+be/PiptD2BsAg==; b2b_order_talk=true; auth.check=B3B6E79576D692134650EAAA904B38A99381377C3AD6103AF2A7E0B853DB7D9A; ws.cms.auth=CfDJ8NOteJ5fo8BHm967u3RJ67-_Fu0B6dkY_Wj7oUDgk7DaVoh1v0FHuksBJjICwUJl7o-mxWSBuw9q1Q2B5y2XM3BRjce_7CQFaqC_62VyCUZZR06PNiyS8UV1z5lBGwq22pNLJ_5QMkhXXqWlCBecYsyQFNJZH9D8KtENVgnmVP5auWJktdS8Z9-xzEJl_VKY1GKiwcONEfuUIGp1CBcXSei1oPJ12ptO9R_3OFPbQl3kV-HLmWeu4R6Xpg4LMVy4rUJjGpgEhzd225qrw3gRCZjOqXCsW7y-XzRRZ-S_HyZhYR5NzgAUPVXF36NcWIMfUvBRIXqD7Kxcm6A8tH5YzACGiA9--EW1zxSS04SPnGJGROPhke09mjz6kAIK7Fn0ORiD7yDN37iK0ZT0y0WzkhgyW5Vg2s98fznBVpBPEmn6IaafF7Wyw0Nay9D_4R1URKuY9yxPBCGhvJRZ9hU_rfu9xxTy-9BtALOUpI-Kj8gw9Wx7POMjlcEBcMbgL3YT09npaSXW0n8XBsO5UUVhsnQ; ws.cms.session-valid=CfDJ8NOteJ5fo8BHm967u3RJ679oCARAJXaDlMm0rDSNiuT0x44ieBgRt7PV7DTpuMHk1PQ0IS3YE6G7uFcfLiIiiAiorWPG753Poaes_jkeG6lEE3B8UUi_Zz727xfv3_melw8rfYTKc_IqmmQ1maRxjVdCHjJtcIA6UlCJ4DH4vKXo; .AspNetCore.Antiforgery.VyLW6ORzMgk=CfDJ8NOteJ5fo8BHm967u3RJ67_imRPApPYOm85Tn23CRY0nKcAqKe_KRaPzKPW_PboCbwZQQkyTatFPl3zpsJ5TCQx-tVS5ld0uauG7j0cnwHr65zZ2-dwy0gfq6f5sV2HM7D7MlJRw1OWyX1j3fL7DDAM; sr_singular=374639ec-4d41-447d-9479-b26146cdd5de"
	req.Header.Set("Cookie", cookie)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	raw := doc.Find(`script#filter-template-LandingDomainId-app-model`).First().Text()
	var p Payload
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return nil, err
	}

	var urls []string

	for _, it := range p.SelectItems {
		fmt.Println(it.Text)
		urls = append(urls, it.Text)
	}

	return urls, nil
}
