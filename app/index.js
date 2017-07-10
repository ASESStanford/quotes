import './style.css';
import _ from 'lodash';
import React from 'react';
import ReactDOM from 'react-dom';


class QuoteList extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      quotes: []
    };

    this.retrieveQuotes();
    this.addQuote = this.addQuote.bind(this);
  }

  retrieveQuotes() {
    return fetch(`/quotes`, {
      method: 'GET'
    })
      .then(response => response.json())
      .then(response => {
        this.setState({
          quotes: response.quotes
        });
      });
  }

  deleteQuote(id) {
    return () => fetch(`/quotes/${id}`, {
      method: 'DELETE'
    })
      .then(response => {
      this.state.quotes = this.state.quotes.filter((quote) => {
        if(quote.id !== id) { return quote; }
      });

      this.setState({
        quotes: this.state.quotes
      });
    });
  }

  addQuote() {
    let quote = document.getElementById('new-quote').value;
    let person= document.getElementById('new-person').value;

    fetch(`/quotes`, {
      method: 'POST',
      headers: {
        'Accept': 'application/json, text/plain, */*',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        quote,
        person
      })
    })
      .then(response => response.text())
      .then(response => {
        this.state.quotes.push({
          id: response,
          quote,
          person
        });

        this.setState({
          quotes: this.state.quotes
        });
        document.getElementById('new-quote').value = '';
        document.getElementById('new-person').value = '';
      });
  }

  render() {
    let list = this.state.quotes.map((quote) => {
      return (
        <Quote quote={quote}
          onDelete={this.deleteQuote(quote.id)}
        ></Quote>
      );
    })

    return (
      <div>
        {list}
        <input type="text" id="new-quote" />
        <input type="text" id="new-person" />
        <button id="add-quote" onClick={this.addQuote}>New</button>
      </div>
    );
  }
}

class Quote extends React.Component {
  constructor(props) {
    super(props);

    this.state = {
      quote: props.quote.quote,
      person: props.quote.person,
      id: props.quote.id
    };

    this.onQuoteChange = this.onQuoteChange.bind(this);
    this.onPersonChange = this.onPersonChange.bind(this);
    this.onUpdate = this.onUpdate.bind(this);
  }

  componentWillReceiveProps(props) {
    this.setState({
      quote: props.quote.quote,
      person: props.quote.person,
      id: props.quote.id
    });
  }

  onQuoteChange(event) {
    this.setState({
      quote: event.target.value
    });
  }

  onPersonChange(event) {
    this.setState({
      person: event.target.value
    });
  }

  onUpdate() {
    return fetch(`/quotes/${this.state.id}`, {
      method: 'POST',
      headers: {
        'Accept': 'application/json, text/plain, */*',
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        id: this.state.id,
        quote: this.state.quote,
        person: this.state.person
      })
    });
  }

  render() {
    let quote = this.props.quote;

    return (
      <div className="quote">
        <input type="text" name="quote" onChange={this.onQuoteChange} value={this.state.quote}/>
        <input type="text" name="person" onChange={this.onPersonChange} value={this.state.person}/>
        <button onClick={this.onUpdate}>Update</button>
        <button onClick={this.props.onDelete}>Delete</button>
      </div>
    )
  }
};

class App extends React.Component {
  render() {
    return (
      <QuoteList></QuoteList>
    )
  }
}

ReactDOM.render(<App />, document.getElementById('container'));
