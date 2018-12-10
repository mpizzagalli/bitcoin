#include <iostream>
#include <vector>
#include <string>
#include <unordered_map>
#include <utility>
#include <map>

using namespace::std;

struct country {
	string id;
	double cnt;
};

map<string, double> innerLat;

void parseCountries() {
	cout << "\t\"country_distribution\":[\n";

	int n;

	cin >> n;

	vector<country> v;
	v.reserve(n);

	int m;
	string id;
	int cnt;
	double ilat;
	int tot =0;

	for (int i=0; i<n; ++i){
		
		cin >> m >> id >> cnt >> ilat;

		if(id != "unknown"){
			/*
				Belarus se excluye por no tener info de latencias
				Georgia e Islandia se incluyen a pesar de tener pocos nodos porque son usados en pools
			*/
			if ((cnt >= 23 && id != "BRL") || id == "GEO" || id == "ISL") { 
				if (id == "USA") {
					v.push_back({"USA-E", ((double)cnt)*0.7});
					innerLat["USA-E"] = 20.0;
					v.push_back({"USA-W", ((double)cnt)*0.3});
					innerLat["USA-W"] = 40.0;
				} else if (id == "CHN") {
					v.push_back({"CHN-N", ((double)cnt)*0.7});
					innerLat["CHN-N"] = 26.0;
					v.push_back({"CHN-S", ((double)cnt)*0.3});
					innerLat["CHN-S"] = 30.0;
				} else {
					v.push_back({id, (double)cnt});
					innerLat[id] = ilat;
				}
			}
			
			tot += cnt;
		}
	}

	double dtot = (double) tot;

	for (int i=0; i<v.size(); ++i){

		double share = v[i].cnt/dtot;
		cout << "\t\t{\n\t\t\t\"id\":\"" << v[i].id << "\",\n\t\t\t\"share\":" << share << ",\n\t\t\t\"inner_latency\":" << (innerLat[v[i].id]/2) << "\n\t\t}";
		if (i<v.size()-1) {
			cout << ',';
		}
		cout << endl;
	}

	cout << "\t],\n";
}

void printLatency(string &a, string &b, double latency) {
	cout << "\t\t{\n";

	cout << "\t\t\t\"a\":\"" << a << "\",\n";
	cout << "\t\t\t\"b\":\"" << b << "\",\n";
	cout << "\t\t\t\"latency_ms\":" << latency << endl;

	cout << "\t\t}";
}

void parseLatencies(){
	cout << "\t\"country_latency\":[\n";

	int n;
	cin >> n;

	vector<string> v;
	v.reserve(n);

	string id;

	int chnni;
	int chnsi;

	for (int i=0; i<n; ++i) {

		cin >> id;
		v.push_back(id);

		if (id=="CHN-N") {
			chnni = i;
		} else if (id=="CHN-S") {
			chnsi = i;
		}

	}

	char slash;

	vector<pair<int, double> > latsAChnN;
	latsAChnN.reserve(n);
	vector<pair<int, double> > latsAChnS;
	latsAChnS.reserve(n);

	double latency;

	for (int i=0; i<n; ++i) {
		cin >> id;
		if (id != "CHN-S" && id != "CHN-N") {
			for (int j=0; j<n; ++j) {
				if (i!=j) {

					cin >> latency;

					double correction = (innerLat[id] + innerLat[v[j]])/4;

					//corregimos la latencia restando una fraccion de las latencias locales
					if (latency-correction<3.5) {
						latency = 3.5;	
					} else {
						latency -= correction;
					}

					if (j == chnni) {
						latsAChnN.push_back(make_pair(i, latency));
					} else if (j == chnsi) {
						latsAChnS.push_back(make_pair(i, latency));
					}

					printLatency(id, v[j], latency);

					cout << ',' << endl;
				} else {
					cin >> slash;
				}
			}
		}
	}

	id = "CHN-N";

	for (int i=0; i<latsAChnN.size(); ++i) {

		printLatency(id, v[latsAChnN[i].first], latsAChnN[i].second);
		cout << ',' << endl;

	}

	id = "CHN-S";

	for (int i=0; i<latsAChnS.size(); ++i) {

		printLatency(id, v[latsAChnS[i].first], latsAChnS[i].second);
		cout  << ',' << endl;
	}

	string idtmp = "CHN-N";

	//dividimos la latencia por 3 por el delay local de chn -s y -n
	printLatency(id, idtmp, (37.782)/3);

	cout << endl;

	cout << "\t],\n";
}

void outputPool(string &id, double share, int n){
	cout << "\t\t{\n";

	cout << "\t\t\t\"id\":\"" << id << "\"," << endl;
	
	cout <<	"\t\t\t\"hp_share\":" << share/100.0 << ',' << endl;

	cout << "\t\t\t\"nodes\":[\n";

	string idp;
	double pd;

	for (int i=1; i<=n; ++i){
		cout << "\t\t\t\t{\n";

		cin >> idp >> pd;

		cout << "\t\t\t\t\t\"country_id\":\"" << idp << "\",\n";

		cout << "\t\t\t\t\t\"pool_share\":" << pd/100.0 << endl;

		cout << "\t\t\t\t}";

		if (i<n)
			cout << ',';
		cout << endl;
	}

	cout << "\t\t\t]\n";

	cout << "\t\t}";
}

void parsePools(){
	cout << "\t\"pool_distribution\":[\n";

	string id;
	double share;
	int n;

	cin >> id >> share >> n;
	
	outputPool(id, share, n);
	while (cin >> id >> share >> n) {
		cout << ",\n";
		outputPool(id, share, n);
	}

	cout << "\n\t]\n";
}

int main(){
	
	cout << "{" << endl;

	parseCountries();

	parseLatencies();

	parsePools();

	cout << "}" << endl;
}