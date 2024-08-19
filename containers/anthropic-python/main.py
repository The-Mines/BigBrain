# main.py
import os
import anthropic
from langchain.llms import Anthropic
from langchain.embeddings import AnthropicEmbeddings
from langchain.text_splitter import CharacterTextSplitter
from langchain.vectorstores import Chroma
from langchain.chains import VectorDBQA

def process_files(directory):
    text_splitter = CharacterTextSplitter(chunk_size=1000, chunk_overlap=0)
    texts = []
    for root, dirs, files in os.walk(directory):
        for file in files:
            file_path = os.path.join(root, file)
            try:
                with open(file_path, 'r') as f:
                    content = f.read()
                    chunks = text_splitter.split_text(content)
                    texts.extend(chunks)
            except Exception as e:
                print(f"Error processing {file_path}: {e}")
    return texts

# Set up LangChain components with Anthropic
anthropic_api_key = os.environ.get("ANTHROPIC_API_KEY")
embeddings = AnthropicEmbeddings(api_key=anthropic_api_key)
texts = process_files("/workspace/source")
db = Chroma.from_texts(texts, embeddings, persist_directory="/workspace/vectordb")
llm = Anthropic(api_key=anthropic_api_key, model="claude-2")
qa = VectorDBQA.from_chain_type(llm=llm, chain_type="stuff", vectorstore=db)

# Analyze codebase
codebase_type = qa.run("What type of codebase is this? Identify the main programming language and any frameworks used.")

print(f"Codebase type: {codebase_type}")

# Save the result for the next task
with open("/workspace/analysis_result.txt", "w") as f:
    f.write(codebase_type)