#include <string>
#include <iostream>

using namespace::std;

int main(){

	unsigned long long t0;
	
	{
		string tmp;
		for (int i=0; i<4; ++i){
			cin >> tmp;
		}
		cin >> t0;
	}

	char id;
	string hash;
	string parent;
	int txAmount;
	unsigned long long timestamp;

	int blocks = 0;

	unsigned long long t1;
	unsigned long long t2;
	unsigned long long t3;

	while (cin >> id >> hash) {
		if (id == '0') {
			cin >> parent >> txAmount;
			blocks++;
		} else if (id == '1') {
			cin >> parent;
			blocks++;
		}

		cin >> timestamp;
		
		if (blocks==576001 && id != '2') {
			cout << hash << endl;
			t1 = timestamp;
		} else if (blocks==577209 && id != '2') {
			cout << hash << endl;
			t2 = timestamp;
		}
	}

	timestamp -= t0;
	timestamp /= 1000000000;
	
	cout << "Processed " << blocks << " blocks in " << timestamp/3600 << ':' << (timestamp%3600)/60 << ':' << ((timestamp%3600)%60) << endl;

	cout << t1 << endl;
	cout << t2 << endl;

	t2 -= t1;
	t2 /= 1000000000;

	cout << "1200 test blocks processed in " << t2/3600 << ':' << (t2%3600)/60 << ':' << ((t2%3600)%60) << endl;


}