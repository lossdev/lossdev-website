from flask import Flask, render_template
import sys
import getopt


app = Flask("lossdev")


@app.route("/")
def index():
    return render_template("index.html")


def main(argv):
    prod = False
    try:
        opts, args = getopt.getopt(argv, "hp", [])
    except getopt.GetoptError:
        print("python app.py [-p production] [-h help]")
        print("Development mode is enabled by default")
        sys.exit(1)
    for opt, arg in opts:
        if opt == "-h":
            print("python app.py [-p production] [-h help]")
            print("Development mode is enabled by default")
            sys.exit()
        elif opt == "-p":
            prod = True
    if prod == True:
        app.run(host="0.0.0.0", port="443", debug=False, ssl_context=("lossdev.pem", "lossdev.key"))
    else:
        app.run(host="0.0.0.0", port="5000", debug=True, ssl_context=("lossdev.pem", "lossdev.key"))


if __name__ == "__main__":
    main(sys.argv[1:])