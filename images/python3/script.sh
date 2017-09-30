for INPUT in $(ls *.in)
do
    echo $INPUT
    python3 source.py < $INPUT
done