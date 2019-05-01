#include <iostream>
#include <random>
#include <algorithm>

using namespace::std;

int main(){

	double l;
	cin >> l;
	
	std::random_device rd;
  std::mt19937 gen(rd());

  std::exponential_distribution<double> d(1/75.0);

  vector<double> v;

  for (int i=0; i<1; ++i) {
    double acum=0;
    for (int j=0; j<1200; j++){

      /*std::random_device rd;
      std::mt19937 gen(rd());

      std::exponential_distribution<double> d(1.0/75.0);*/
      //double secondsToWait = d(gen);

      //secondsToWait *= 75;

      acum += d(gen);

      v.push_back(acum);
    }
  	 
  }

  sort(v.begin(), v.end());

  double tot = v[0];

  for (int i=1; i<v.size();i++) {

    tot += (v[i]-v[i-1]);

  }

    cout << endl << "Tiempo promedio de espera: "<< tot/v.size() << endl;

}