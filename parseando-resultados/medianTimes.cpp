#include <iostream>
#include <string>
#include <vector>

using namespace::std;

vector<double> times;

void printPercentileTimes(string &tmp){

	double time;

	tmp.push_back(',');

	while (tmp.size()>1) {
		cin >> tmp;
		cin >> time;
		times.push_back(time);
		cin >> tmp;
	}
	if (times.size()>=10) {
		int interval = (times.size()+5)/10;
		double total=0.0;
		for (int j=1;j<times.size();j++) {
			total += times[j];
			if (j%interval==0) {
				cout << times[j] << ' ';
			}
		}
		cout << times.back() << endl << total/((double)(times.size()-1)) << endl;
	}
	times.clear();
}

int main(){
	
	int i = 0;
	string hash;
	string parent;
	int nTx;
	string tmp = "s,";

	while (i<1200) {
		cin >> i >> hash >> parent >> nTx;
		cout << i << ' ' << hash << ' ' << parent << ' ' << nTx << endl;
		
		printPercentileTimes(tmp);
		printPercentileTimes(tmp);
	}
}