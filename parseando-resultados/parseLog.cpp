#include <string>
#include <iostream>

using namespace::std;

int main(){

	unsigned long long t0;
	
	// Ignore the prefix "Starting bitcoin client at" but get the node starting time
	{
		string tmp;
		for (int i=0; i<4; ++i){
			cin >> tmp;
		}
		cin >> t0;
	}

	char id;
	string hash;				  // Hash del bloque
	string parent;				  // Hash del padre del bloque
	int txAmount;				  // Cantidad de tx del bloque
	unsigned long long timestamp; // Timestamp del bloque

	int blocks = 0;

	unsigned long long t1;
	unsigned long long t2;
	unsigned long long t3;

	while (cin >> id >> hash) {
		// We calculate the numbers of blocks produced in the whole network
		if (id == '0') {
			cin >> parent >> txAmount;
			blocks++;
		} else if (id == '1') {
			cin >> parent;
			blocks++;
		}

		cin >> timestamp;
		
		// We get
		//	t1=timestamp of block 576001
		// 	t2=timestamp of block 577209 
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
	
	// We output the amount of block processed and the time it took for them
	cout << "Processed " << blocks << " blocks in " << timestamp/3600 << ':' << (timestamp%3600)/60 << ':' << ((timestamp%3600)%60) << endl;

	// We output the timestamp of the selected blocks
	cout << t1 << endl;
	cout << t2 << endl;

	t2 -= t1;
	t2 /= 1000000000;

	// We output the time difference between the selected blocks
	cout << "1200 test blocks processed in " << t2/3600 << ':' << (t2%3600)/60 << ':' << ((t2%3600)%60) << endl;


}