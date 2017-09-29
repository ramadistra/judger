import requests
import json

from flask import Flask, render_template, request

WEBSERIVCE_URL = "http://localhost:8000"
app = Flask(__name__)


@app.route('/')
def home():
    return render_template('index.html')


@app.route('/submit', methods=['POST'])
def submit():
    script = request.form['message']
    result, accepted = verify(script, "a/")
    return render_template('result.html', cases=result, accepted=accepted)


@app.route('/questions/<id>', methods=['GET'])
def question(id):
    return render_template('questions/%s.html'%id)


def verify(script, question):
    n = int(open(question + "cases.txt", 'r').read())
    accepted = True
    result = []
    
    for i in range(1, n + 1):
        stdin = open("{}{}.in".format(question, i)).read()
        expected_output = open("{}{}.out".format(question, i)).read()
        response_http = requests.post(WEBSERIVCE_URL + "/python3", data={
            'source':script, 
            'stdin':stdin,
            'timeout':500,
            })
            
        response = json.loads(response_http.text)
        if response['stdout'] == expected_output:
            result.append("correct")
        else:
            status = response['status']
            result.append('wrong :(' if status == 'OK' else status)
            accepted = False

    return result, accepted


if __name__ == "__main__":
    app.run(debug=True, port=5050)
