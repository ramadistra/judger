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
    result, inputs = [], []
    
    for i in range(1, n + 1):
        stdin = open("{}{}.in".format(question, i)).read()
        inputs.append(stdin)
        
    data = json.dumps({'source':script, 'stdin':inputs, 'timeout':500})
    response_http = requests.post(WEBSERIVCE_URL + "/python3", data=data)
    response = json.loads(response_http.text)
    stdouts = response['stdout']
    outputs = parse_stdouts(stdouts, n)
    answers = [open("{}{}.out".format(question, i+1)).read() for i in range(n)]

    for expected, got in zip(answers, outputs):
        if expected.strip() == got.strip():
            result.append("correct")
        else:
            status = response['status']
            result.append('wrong :(' if status == 'OK' else status)
            accepted = False

    return result, accepted


def parse_stdouts(output, n):
    i = 0
    stdouts = ['']*n
    for line in output.split('\n'):
        if i > n:
            break
        if line == "%d.in" % (i+1):
            i += 1
        else:
            stdouts[i-1] += line + '\n'

    return stdouts


if __name__ == "__main__":
    app.run(debug=True, port=5050)
