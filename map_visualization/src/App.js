import React from 'react';
import logo from './logo.svg';
import './App.css';
import LeafletMap from './components/LeafletMap.js';
import TimeController from './components/TimeController';

class App extends React.Component {

  constructor(props) {
    super(props);
    this.state = {
      runSimulationRev: null
    }
  }

  runSimulation(timestamp) {
    let rev = this.state.runSimulationRev ? this.state.runSimulationRev + 1 : 1;
    
    this.setState({
      runSimulationRev: rev,
      timestamp: timestamp
    })
  }

  render() {
    return (
      <div>
        <LeafletMap runSimulationRev={this.state.runSimulationRev} timestamp={this.state.timestamp} />
        <TimeController notifyStartSimulation={(ts) => this.runSimulation(ts)} />
      </div>
    );
  }
}


export default App;
